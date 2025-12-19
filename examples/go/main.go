package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	DefaultURL = "https://latexlite.com"
	DemoAPIKey = "demo-key-1234567890abcdef"
)

func escape(s string) string {
	replacer := strings.NewReplacer(
		"&", `\&`,
		"%", `\%`,
		"$", `\$`,
		"#", `\#`,
		"_", `\_`,
		"{", `\{`,
		"}", `\}`,
		"~", `\textasciitilde{}`,
		"^", `\^{}`,
		"\\", `\textbackslash{}`,
	)
	return replacer.Replace(s)
}

func main() {
	// Get API credentials from environment or use defaults
	apiURL := getEnv("LATEX_API_URL", DefaultURL)
	apiKey := getEnv("LATEX_API_KEY", DemoAPIKey)

	fmt.Printf("🚀 LaTeX Lite API Go Client\n")
	fmt.Printf("API URL: %s\n", apiURL)
	fmt.Printf("API Key: %s...\n\n", apiKey[:10])

	// Escape LaTeX special characters in invoice item descriptions
	for _, item := range invoiceData["Items"].([]map[string]interface{}) {
		item["Description"] = escape(item["Description"].(string))
	}

	client := NewLatexClient(apiURL, apiKey)

	// Example 1: Simple document
	fmt.Println("📄 Example 1: Simple Document")
	simpleJob, err := client.CreateAndWait(
		`\documentclass{article}\begin{document}\title{ {{.Title}} }\author{ {{.Author}} }\maketitle

{{.Content}}\end{document}`,
		map[string]interface{}{
			"Title":   "My First PDF",
			"Author":  "Go Client",
			"Content": "This PDF was generated using the LaTeX Lite API!",
		},
		"simple.pdf",
	)
	if err != nil {
		log.Printf("❌ Simple example failed: %v", err)
	} else {
		fmt.Printf("✅ Success: %s\n\n", simpleJob.ID)
	}

	// Example 2: Invoice
	fmt.Println("🧾 Example 2: Invoice")
	invoiceJob, err := client.CreateAndWait(invoiceTemplate, invoiceData, "invoice.pdf")
	if err != nil {
		log.Printf("❌ Invoice example failed: %v", err)
	} else {
		fmt.Printf("✅ Success: %s\n\n", invoiceJob.ID)
	}

	fmt.Println("🎉 All examples complete! Check the generated PDF files.")
}

// LaTeX client implementation
type LatexClient struct {
	BaseURL string
	APIKey  string
	Client  *http.Client
}

type RenderRequest struct {
	Template string                 `json:"template"`
	Data     map[string]interface{} `json:"data"`
}

type RenderJob struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Error     *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
	Log string `json:"log,omitempty"` // <-- Add this line
}

type APIResponse struct {
	Success bool       `json:"success"`
	Data    *RenderJob `json:"data,omitempty"`
	Error   *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func NewLatexClient(baseURL, apiKey string) *LatexClient {
	return &LatexClient{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Client:  &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *LatexClient) CreateAndWait(template string, data map[string]interface{}, filename string) (*RenderJob, error) {
	// Create job
	job, err := c.CreateRender(template, data)
	if err != nil {
		return nil, fmt.Errorf("create render: %w", err)
	}
	if job == nil {
		return nil, fmt.Errorf("create render: job is nil (API error or malformed response)")
	}
	fmt.Printf("  Job created: %s\n", job.ID)

	// Wait for completion
	job, err = c.WaitForCompletion(job.ID, 60*time.Second)
	if err != nil {
		return nil, fmt.Errorf("wait for completion: %w", err)
	}

	if job.Status != "succeeded" {
		errMsg := "unknown error"
		if job.Error != nil {
			errMsg = job.Error.Message
		}
		// Show LaTeX log if present
		if job.Log != "" {
			errMsg += "\n\nLaTeX log:\n" + job.Log
		}
		return nil, fmt.Errorf("job failed: %s", errMsg)
	}

	// Download PDF
	err = c.DownloadPDF(job.ID, filename)
	if err != nil {
		return nil, fmt.Errorf("download PDF: %w", err)
	}

	fmt.Printf("  PDF downloaded: %s\n", filename)
	return job, nil
}

func (c *LatexClient) CreateRender(template string, data map[string]interface{}) (*RenderJob, error) {
	req := RenderRequest{Template: template, Data: data}
	body, _ := json.Marshal(req)

	httpReq, err := http.NewRequest("POST", c.BaseURL+"/v1/renders", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	json.NewDecoder(resp.Body).Decode(&apiResp)

	if !apiResp.Success {
		return nil, fmt.Errorf("API error: %s", apiResp.Error.Message)
	}

	return apiResp.Data, nil
}

func (c *LatexClient) WaitForCompletion(jobID string, timeout time.Duration) (*RenderJob, error) {
	start := time.Now()
	for {
		job, err := c.GetRender(jobID)
		if err != nil {
			return nil, err
		}

		switch job.Status {
		case "succeeded", "failed":
			return job, nil
		}

		if time.Since(start) > timeout {
			return job, fmt.Errorf("timeout waiting for job completion")
		}

		time.Sleep(2 * time.Second)
		fmt.Print(".")
	}
}

func (c *LatexClient) GetRender(jobID string) (*RenderJob, error) {
	httpReq, err := http.NewRequest("GET", c.BaseURL+"/v1/renders/"+jobID, nil)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.Client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	json.NewDecoder(resp.Body).Decode(&apiResp)

	if !apiResp.Success {
		return nil, fmt.Errorf("API error: %s", apiResp.Error.Message)
	}

	return apiResp.Data, nil
}

func (c *LatexClient) DownloadPDF(jobID, filename string) error {
	httpReq, err := http.NewRequest("GET", c.BaseURL+"/v1/renders/"+jobID+"/pdf", nil)
	if err != nil {
		return err
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.Client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to download PDF: %s", resp.Status)
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

const invoiceTemplate = `\documentclass{article}
\usepackage[margin=1in]{geometry}
\begin{document}
\begin{center}{\Large \textbf{INVOICE}}\end{center}
\vspace{1em}
\noindent\textbf{Invoice \#:} {{.InvoiceNumber}} \\
\textbf{Date:} {{.Date}}
\vspace{1em}
\noindent\textbf{Bill To:} \\{{.CustomerName}} \\{{.CustomerAddress}}
\vspace{2em}
\begin{tabular}{|l|r|}
\hline
\textbf{Description} & \textbf{Amount} \\
\hline
{{range .Items}}{{.Description}} & \${{.Amount}} \\
\hline
{{end}}
\multicolumn{1}{|r|}{\textbf{Total:}} & \textbf{\${{.Total}}} \\
\hline
\end{tabular}
\vspace{2em}
Thank you for your business!
\end{document}`

var invoiceData = map[string]interface{}{
	"InvoiceNumber":   "INV-GO-001",
	"Date":            "December 14, 2025",
	"CustomerName":    "Tech Startup Inc",
	"CustomerAddress": "456 Innovation Drive\\\\San Francisco, CA 94105",
	"Items": []map[string]interface{}{
		{"Description": "LaTeX API Integration", "Amount": "1500.00"},
		{"Description": "Custom Templates", "Amount": "800.00"},
		{"Description": "Support & Training", "Amount": "700.00"},
	},
	"Total": "3000.00",
}
