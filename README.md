# LaTeXLite API - Quickstart Guide

Generate professional PDFs from LaTeX templates with a simple REST API.

## 🚀 Quick Start

### 1. Sign Up for an API Key and set Environment Variables

Visit [latexlite.com/get-demo-key](https://latexlite.com/get-demo-key) for a free demo API key. Export it as an environment variable for use, along with the base API URL:

```bash
# Demo API key (rate limited)
export API_KEY="<your-api-key>"
export BASE_URL="https://latexlite.com"
```

## When to use Sync vs Async

- **Sync (`/v1/renders-sync`)**: Best for **small, single** renders where you want the PDF immediately (e.g. “generate one PDF and download it now”). No polling required.
- **Async (`/v1/renders`)**: Best for **longer/heavier** renders and **parallel** workloads (many PDFs). Create jobs, poll for completion, then download. More reliable for work that may exceed short request timeouts.

## Endpoints

### Render PDF (synchronous)

```bash
POST /v1/renders-sync
```

- Returns **`application/pdf`** by default (recommended for `curl -o out.pdf`)
- If you set **`Accept: application/json`**, it returns a JSON envelope containing `pdf_base64`
- On error (e.g. invalid API key), it returns a JSON error body with a non-2xx status code.

## Async (job-based) rendering

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
    "expires_at": "2024-01-15T11:30:00Z"
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
    "expires_at": "2024-01-15T11:30:00Z",
    "pdf_url": "/v1/renders/job_1234567890/pdf"
  }
}
```

### Download PDF

```bash
GET /v1/renders/{id}/pdf
```

Returns the compiled PDF file when status is "succeeded".

## Math (synchronous)

### Render LaTeX math (sync)

```bash
POST /v1/math-sync
```

- Returns a PNG image by default (recommended for `curl -o equation.png`)
- If you set **`Accept: application/json`**, it returns a JSON response (useful for programmatic handling)

## Job Status Values

- `queued`: Job is waiting to be processed
- `running`: Job is currently being compiled
- `succeeded`: PDF generated successfully
- `failed`: Compilation failed (check error field)
- `expired`: Job expired (24h TTL)

## Request Limits

### Async (/v1/renders)
- Max template size: 1MB
- Max compilation time: 20 seconds
- Max PDF size: 20MB

### Sync (/v1/renders-sync)
- Intended for short renders (may time out for heavy workloads); for large/slow documents use async.

## Example Usage

```bash
# Sync: Render and save PDF directly (recommended for single small jobs)
curl -sS -X POST "${BASE_URL}/v1/renders-sync" \
  -H "Authorization: Bearer ${API_KEY}" \
  -H "Content-Type: application/json" \
  -o out.pdf \
  -d '{
    "template": "\\documentclass{article}\n\\begin{document}\nHello, [[.Who]]!\n\\end{document}",
    "data": { "Who": "world" }
  }'

# Sync: Render and return JSON (base64 PDF) for programmatic handling
curl -sS -X POST "${BASE_URL}/v1/renders-sync" \
  -H "Authorization: Bearer ${API_KEY}" \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -d '{
    "template": "\\documentclass{article}\n\\begin{document}\nHello, [[.Who]]!\n\\end{document}",
    "data": { "Who": "world" }
  }'

# Async: Simple LaTeX without templating
curl -X POST "${BASE_URL}/v1/renders" \
  -H "Authorization: Bearer ${API_KEY}" \
  -H "Content-Type: application/json" \
  -d '{
    "template": "\\documentclass{article}\\begin{document}Hello World!\\end{document}"
  }'

# Async: LaTeX with Go templating and [[ ]] delimiters
curl -X POST "${BASE_URL}/v1/renders" \
  -H "Authorization: Bearer ${API_KEY}" \
  -H "Content-Type: application/json" \
  -d '{
    "template": "\\documentclass{article}\\begin{document}Invoice for [[.CustomerName]] \\\\ Amount: \\$[[.Amount]]\\end{document}",
    "data": {
      "CustomerName": "John Doe",
      "Amount": "1250.00"
    }
  }'

# Async: Check status
curl -H "Authorization: Bearer ${API_KEY}" \
  "${BASE_URL}/v1/renders/job_1234567890"

# Async: Download PDF when ready
curl -H "Authorization: Bearer ${API_KEY}" \
  "${BASE_URL}/v1/renders/job_1234567890/pdf" \
  -o output.pdf

# Sync math: render LaTeX math and save as PNG
curl -sS -X POST "${BASE_URL}/v1/math-sync" \
  -H "Authorization: Bearer ${API_KEY}" \
  -H "Content-Type: application/json" \
  -o equation.png \
  -d '{
    "math": "$\\int_0^1 x^2 \\, dx = \\frac{1}{3}$"
  }'

# Sync math: request JSON response
curl -sS -X POST "${BASE_URL}/v1/math-sync" \
  -H "Authorization: Bearer ${API_KEY}" \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -d '{
    "math": "$E = mc^2$"
  }'
```

> Note: In JSON, backslashes must be escaped. That’s why LaTeX commands use `\\int`, `\\frac`, and `\\,` inside the JSON string.

## Error Handling

| Status | Meaning |
|--------|---------|
| `401` | Invalid API key |
| `429` | Rate limit exceeded |
| `400` | Invalid template or data |
| `409` | PDF not ready (still processing) |

## Best Practices

1. **Pick sync vs async appropriately** - Sync for small single renders; async for heavier/parallel workloads
2. **Escape LaTeX characters** - `\` must become `\\` in JSON strings. The Go quickstart code can escape for you.
3. **Poll for completion (async)** - Check status every 2-5 seconds
4. **Cache PDFs** - Jobs expire after 24 hours
5. **Handle rate limits** - Respect the `X-RateLimit-*` headers

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
