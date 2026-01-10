package main

import (
	"bytes"
	"encoding/base64"
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
	BaseURL    = "https://latexlite.com"
	DemoAPIKey = "<your-api-key>"
)

func main() {
	// Get API credentials from environment or use defaults
	baseURL := getEnv("BASE_URL", BaseURL)
	apiKey := getEnv("API_KEY", DemoAPIKey)

	fmt.Printf("🚀 LaTeX Lite API Go Client\n")
	fmt.Printf("API URL: %s\n", baseURL)
	fmt.Printf("API Key: %s...\n\n", previewKey(apiKey, 10))

	// Escape LaTeX special characters in invoice item descriptions
	for _, item := range invoiceData["Items"].([]map[string]interface{}) {
		item["Description"] = escape(item["Description"].(string))
	}

	client := NewLatexClient(baseURL, apiKey)

	// Example 1: Sync render (no polling)
	fmt.Println("⚡ Example 1: Sync Render (renders-sync)")
	if err := client.RenderSyncToFile(
		`\documentclass{article}\begin{document}Hello, [[.Who]]!\end{document}`,
		map[string]interface{}{"Who": "sync world"},
		"sync.pdf",
	); err != nil {
		log.Printf("Sync example failed: %v", err)
	} else {
		fmt.Printf("  PDF downloaded: %s\n\n", "sync.pdf")
	}

	// Example 2: Simple document
	fmt.Println("📄 Example 2: Simple Document")
	simpleJob, err := client.CreateAndWait(
		`\documentclass{article}\begin{document}\title{ [[.Title]] }\author{ [[.Author]] }\maketitle

[[.Content]] \end{document}`,
		map[string]interface{}{
			"Title":   "My First PDF",
			"Author":  "Go Client",
			"Content": "This PDF was generated using the LaTeX Lite API!",
		},
		"simple.pdf",
	)
	if err != nil {
		log.Printf("Simple example failed: %v", err)
	} else {
		fmt.Printf("Success: %s\n\n", simpleJob.ID)
	}

	// Example 3: Invoice
	fmt.Println("Example 3: Invoice")
	invoiceJob, err := client.CreateAndWait(invoiceTemplate, invoiceData, "invoice.pdf")
	if err != nil {
		log.Printf("Invoice example failed: %v", err)
	} else {
		fmt.Printf("Success: %s\n\n", invoiceJob.ID)
	}

	fmt.Println("All examples complete! Check the generated PDF files.")
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
	Log string `json:"log,omitempty"`
}

type APIResponse struct {
	Success bool       `json:"success"`
	Data    *RenderJob `json:"data,omitempty"`
	Error   *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type SyncRenderResponse struct {
	Success bool `json:"success"`
	Data    *struct {
		PDFBase64 string `json:"pdf_base64"`
	} `json:"data,omitempty"`
	Error *struct {
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

// RenderSyncToFile calls POST /v1/renders-sync and writes the PDF to filename.
// It requests application/pdf (recommended). If the server responds with JSON,
// it will parse pdf_base64 and write it to filename.
func (c *LatexClient) RenderSyncToFile(template string, data map[string]interface{}, filename string) error {
	req := RenderRequest{Template: template, Data: data}
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest("POST", c.BaseURL+"/v1/renders-sync", bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/pdf")

	resp, err := c.Client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Non-2xx: API returns JSON error body.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := readAPIErrorMessage(resp.Body)
		if msg == "" {
			msg = resp.Status
		}
		return fmt.Errorf("sync render failed: %s", msg)
	}

	ct := resp.Header.Get("Content-Type")
	// If we got PDF bytes, stream to file.
	if strings.HasPrefix(ct, "application/pdf") {
		return writeBodyToFile(resp.Body, filename)
	}

	// Otherwise attempt JSON envelope with pdf_base64.
	var apiResp SyncRenderResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("sync render: expected PDF but got %q and JSON decode failed: %w", ct, err)
	}
	if !apiResp.Success {
		if apiResp.Error != nil && apiResp.Error.Message != "" {
			return fmt.Errorf("sync render API error: %s", apiResp.Error.Message)
		}
		return fmt.Errorf("sync render API error: unknown error")
	}
	if apiResp.Data == nil || apiResp.Data.PDFBase64 == "" {
		return fmt.Errorf("sync render: missing pdf_base64 in response")
	}

	pdfBytes, err := base64.StdEncoding.DecodeString(apiResp.Data.PDFBase64)
	if err != nil {
		return fmt.Errorf("sync render: decode pdf_base64: %w", err)
	}
	return os.WriteFile(filename, pdfBytes, 0o644)
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
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if !apiResp.Success {
		if apiResp.Error != nil && apiResp.Error.Message != "" {
			return nil, fmt.Errorf("API error: %s", apiResp.Error.Message)
		}
		return nil, fmt.Errorf("API error: unknown error")
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
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if !apiResp.Success {
		if apiResp.Error != nil && apiResp.Error.Message != "" {
			return nil, fmt.Errorf("API error: %s", apiResp.Error.Message)
		}
		return nil, fmt.Errorf("API error: unknown error")
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

func previewKey(s string, n int) string {
	if n <= 0 {
		return ""
	}
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func writeBodyToFile(r io.Reader, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}

func readAPIErrorMessage(r io.Reader) string {
	var e struct {
		Success bool `json:"success"`
		Error   *struct {
			Message string `json:"message"`
		} `json:"error,omitempty"`
	}
	b, _ := io.ReadAll(r)
	if len(b) == 0 {
		return ""
	}
	if err := json.Unmarshal(b, &e); err == nil {
		if e.Error != nil {
			return e.Error.Message
		}
	}
	// fallback: include raw body (trimmed) if not JSON
	s := strings.TrimSpace(string(b))
	if len(s) > 500 {
		s = s[:500] + "…"
	}
	return s
}

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

const invoiceTemplate = `\documentclass{article}
\usepackage[margin=1in]{geometry}
\begin{document}
\begin{center}{\Large \textbf{INVOICE}}\end{center}
\vspace{1em}

\noindent\textbf{Invoice \#:} [[.InvoiceNumber]] \\
\textbf{Date:} [[.Date]]

\vspace{1em}
\noindent\textbf{Bill To:} \\
[[.CustomerName]] \\
[[.CustomerAddress]]

\vspace{2em}

% --- Table starts here ---
\vspace{1em}
\begin{tabular}{|p{8cm}|r|}
\hline
\textbf{Description} & \textbf{Amount} \\
\hline
[[range .Items]] [[.Description]] & \$[[.Amount]] \\
\hline
[[end]]
\textbf{Total:} & \textbf{\$[[.Total]]} \\
\hline
\end{tabular}
\vspace{2em}
% --- Table ends here ---

\vspace{2em}

\noindent Thank you for your business!

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
