# LaTeXLite API - Quickstart Guide

Generate professional PDFs from LaTeX templates with a simple REST API.

## Quick Start

### 1. Sign Up for an API Key and set Environment Variables

Visit [latexlite.com/get-demo-key](https://latexlite.com/get-demo-key) for a free demo API key. Export it as an environment variable for use, along with the base API URL:

```bash
# Demo API key (rate limited)
export API_KEY="<your-api-key>"
export BASE_URL="https://latexlite.com"
```

## When to use Sync vs Async

- **Sync (`/v1/renders-sync`)**: Best for **small, single** renders where you want the PDF immediately (e.g. "generate one PDF and download it now"). No polling required.
- **Async (`/v1/renders`)**: Best for **longer/heavier** renders and **parallel** workloads (many PDFs). Create jobs, poll for completion, then download. More reliable for work that may exceed short request timeouts.

## Endpoints

### Synchronous PDF Render

```bash
POST /v1/renders-sync
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

**Response (application/pdf):**
Returns PDF binary directly (default behavior). Use `curl -o output.pdf` to save.

**Response (application/json):**
If you set `Accept: application/json`:
```json
{
  "success": true,
  "data": {
    "content_type": "application/pdf",
    "pdf_base64": "JVBERi0xLjQKJeLjz9MKMyAwIG9iaiA8PC..."
  }
}
```

**Error Response:**
```json
{
  "success": false,
  "error": {
    "message": "LaTeX compilation failed: ! Undefined control sequence.",
    "line": 0
  }
}
```

### Async (job-based) rendering

#### Create Render Job

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

#### Get Render Status

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

#### Download PDF

```bash
GET /v1/renders/{id}/pdf
```

Returns the compiled PDF file when status is "succeeded".

### Math (synchronous)

#### Render LaTeX math (sync)

```bash
POST /v1/math-sync
```

**Request:**
```json
{
  "math": "$\\int_0^1 x^2 \\, dx = \\frac{1}{3}$"
}
```

**Response (image/png):**
Returns PNG binary directly (default behavior). Use `curl -o equation.png` to save.

**Response (application/json):**
If you set `Accept: application/json`:
```json
{
  "success": true,
  "data": {
    "content_type": "image/png",
    "png_base64": "iVBORw0KGgoAAAANSUhEUgAA..."
  }
}
```

**Error Response:**
```json
{
  "success": false,
  "error": {
    "message": "math must start and end with $ (or $$)",
    "line": 0
  }
}
```

## Job Status Values

- `queued`: Job is waiting to be processed
- `running`: Job is currently being compiled
- `succeeded`: PDF generated successfully
- `failed`: Compilation failed (check error field)
- `expired`: Job expired (1h TTL)

## HTTP Status Codes

| Status | Meaning |
|--------|---------|
| `200 OK` | Request successful |
| `201 Created` | Render job created successfully |
| `400 Bad Request` | Invalid template or data |
| `401 Unauthorized` | Invalid API key |
| `404 Not Found` | Job not found |
| `408 Request Timeout` | Sync render timed out (use async for larger documents) |
| `409 Conflict` | PDF not ready (still processing) |
| `422 Unprocessable Entity` | LaTeX compilation failed |
| `429 Too Many Requests` | Rate limit exceeded |
| `502 Bad Gateway` | Renderer service error |
| `503 Service Unavailable` | Renderer not configured |

## Request Limits

### Async (/v1/renders)
- Max template size: 200KB
- Max compilation time: 20 seconds
- Max PDF size: 20MB
- Job expiry: 1 hour

### Sync (/v1/renders-sync)
- Max body size: 1MB (configurable via `SYNC_RENDER_MAX_BODY_BYTES`)
- Timeout: 8 seconds (configurable via `SYNC_RENDER_TIMEOUT_SECONDS`)
- Intended for short renders; may time out for heavy workloads

### Math Sync (/v1/math-sync)
- Max body size: 64KB (configurable via `MATH_SYNC_MAX_BODY_BYTES`)
- Max math input: 32KB
- Timeout: 8 seconds (configurable via `MATH_SYNC_TIMEOUT_SECONDS`)

## Example Usage

```bash
# Sync: Render and save PDF directly (recommended for single small jobs)
curl -sS -X POST "${BASE_URL}/v1/renders-sync" \
  -H "Authorization: Bearer ${API_KEY}" \
  -H "Accept: application/pdf" \
  -H "Content-Type: application/json" \
  -o out.pdf \
  -d '{
    "template": "\\documentclass{article}\n\\begin{document}\nHello, [[.Who]]!\n\\end{document}"
  }'

# Sync: Render and return JSON (base64 PDF) with dynamic input data for programmatic handling
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

> Note: In JSON, backslashes must be escaped. That's why LaTeX commands use `\\int`, `\\frac`, and `\\,` inside the JSON string. However, \n (newline) and \t (tab) are JSON escape sequences and should remain single backslash.
{

## Error Handling

All API responses follow this structure:

**Success:**
```json
{
  "success": true,
  "data": { ... }
}
```

**Error:**
```json
{
  "success": false,
  "error": {
    "message": "Error description",
    "line": 0
  }
}
```

Common error messages:
- `"Missing or invalid Authorization header"` (401)
- `"Invalid API key"` (401)
- `"API key rate limit exceeded. Try again in 1 minute."` (429)
- `"IP rate limit exceeded. Maximum 50 requests per minute per IP."` (429)
- `"missing required field: template"` (400)
- `"template too large (max 200KB)"` (400)
- `"LaTeX compilation failed: ...LaTeX error details..."` (422)
- `"render timed out. Use async /v1/renders for larger documents."` (408)
- `"PDF not ready"` (409)
- `"Job not found"` (404)

## Best Practices

1. **Pick sync vs async appropriately** - Sync for small single renders; async for heavier/parallel workloads
2. **Escape LaTeX characters in JSON** - Backslashes must be doubled: `\` becomes `\\` for LaTeX commands like `\\int`, `\\frac`, `\\$`, etc. However, `\n` (newline) and `\t` (tab) are JSON escape sequences and should remain single backslash.
   ```json
   {
     "template": "\\documentclass{article}\n\\begin{document}\nHello \\textbf{World}!\n\\end{document}"
   }
   ```
3. **Poll for completion (async)** - Check status every 2-5 seconds
4. **Cache PDFs** - Async jobs expire after 1 hour
5. **Handle rate limits** - Respect the `X-RateLimit-*` headers:
   - `X-RateLimit-Limit`: Maximum requests allowed per minute
   - `X-RateLimit-Remaining`: Requests remaining in current window
   - `X-RateLimit-Reset`: Unix timestamp when the limit resets
6. **Math input format** - Math strings must start and end with `$` or `$$`
7. **Error handling** - Always check `success` field and handle non-200 status codes

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.