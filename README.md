# LaTeXLite API - Quickstart Guide

Generate professional PDFs from LaTeX templates with a simple REST API.

## 🚀 Quick Start

### 1. Sign Up for an API Key

Visit [latexlite.com/get-demo-key](https://latexlite.com/get-demo-key) for a free demo API key. Use this in place of `<your-api-key>` in the examples below.

### 2. Get API Access
```bash
# Demo API key (rate limited)
export LATEX_API_KEY="<your-api-key>"
export LATEX_API_URL="https://latexlite.com"
```

## Endpoints

### Create Render Job

```bash
POST /v1/renders
```

**Request:**
```json
{
  "template": "\\documentclass{article}\n\\begin{document}\nHello [[.Name]]!\n\\end{document}",
  "data": {
    "Name": "World"
  }
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "job_1234567890",
    "status": "queued",
    "created_at": "2024-01-15T10:30:00Z",
    "expires_at": "2024-01-16T10:30:00Z"
  }
}
```

### Get Render Status

```bash
GET /v1/renders/{id}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "job_1234567890",
    "status": "succeeded",
    "created_at": "2024-01-15T10:30:00Z",
    "expires_at": "2024-01-16T10:30:00Z",
    "pdf_url": "/v1/renders/job_1234567890/pdf"
  }
}
```

### Download PDF

```bash
GET /v1/renders/{id}/pdf
```

Returns the compiled PDF file when status is "succeeded".

## Job Status Values

- `queued`: Job is waiting to be processed
- `running`: Job is currently being compiled
- `succeeded`: PDF generated successfully
- `failed`: Compilation failed (check error field)
- `expired`: Job expired (24h TTL)

## Request Limits

- Max template size: 1MB
- Max compilation time: 20 seconds
- Max PDF size: 20MB

## Example Usage

```bash
# Simple LaTeX without templating
curl -X POST https://latexlite.com/v1/renders \
  -H "Authorization: Bearer <your-api-key>" \
  -H "Content-Type: application/json" \
  -d '{
    "template": "\\documentclass{article}\\begin{document}Hello World!\\end{document}",
    "data": {}
  }'

# LaTeX with Go templating and [[ ]] delimiters
curl -X POST https://latexlite.com/v1/renders \
  -H "Authorization: Bearer <your-api-key>" \
  -H "Content-Type: application/json" \
  -d '{
    "template": "\\documentclass{article}\\begin{document}Invoice for [[.CustomerName]] \\\\ Amount: \\$[[.Amount]]\\end{document}",
    "data": {
      "CustomerName": "John Doe",
      "Amount": "1250.00"
    }
  }'

# Check status
curl -H "Authorization: Bearer <your-api-key>" \
  https://latexlite.com/v1/renders/job_1234567890

# Download PDF when ready
curl -H "Authorization: Bearer <your-api-key>" \
  https://latexlite.com/v1/renders/job_1234567890/pdf \
  -o output.pdf
```


## Error Handling

| Status | Meaning |
|--------|---------|
| `401` | Invalid API key |
| `429` | Rate limit exceeded |
| `400` | Invalid template or data |
| `409` | PDF not ready (still processing) |

## Best Practices

1. **Handle async processing** - Jobs are queued and processed in background
2. **Escape LaTeX characters** - `\` becomes `\\` in JSON strings  
3. **Poll for completion** - Check status every 2-5 seconds
4. **Cache PDFs** - Jobs expire after 24 hours
5. **Handle rate limits** - Respect the `X-RateLimit-*` headers

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
