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

## Endpoints

### PDF Rendering

```bash
POST /v1/renders-sync
```

**Behavior:**
- Returns **`application/pdf`** by default (recommended for `curl -o out.pdf`)
- If you set **`Accept: application/json`**, it returns JSON success with `pdf_base64`
- On error (e.g. invalid API key), returns JSON error body with a non-2xx status

Intended for short renders (defaults to ~**8 seconds** timeout). Heavy workloads may fail with `408 Request Timeout`.

**Supported Content-Type values:**

| `Content-Type` | Use case | Template source | Data source |
|---|---|---|---|
| `application/json` | Inline template string with optional data | `template` field in JSON body | `data` field in JSON body (optional) |
| `text/plain` / `text/x-tex` / `application/x-tex` | Raw `.tex` file without data injection | Raw request body | None |
| `multipart/form-data` | File upload with data injection | `template` file part | `data` form field (inline JSON string) |

**Request (JSON with template string):**
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

### Math Rendering

```bash
POST /v1/math-sync
```

**Behavior:**
- Request JSON body must include a `math` string that **starts and ends with `$`** (or `$$`)
- Returns **`image/png`** by default
- If you set **`Accept: application/json`**, it returns JSON success with `png_base64`

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

## HTTP Status Codes

| Status | Meaning |
|--------|---------|
| `200 OK` | Request successful |
| `400 Bad Request` | Invalid template or data |
| `401 Unauthorized` | Invalid API key |
| `408 Request Timeout` | Render timed out (try optimizing your template) |
| `422 Unprocessable Entity` | LaTeX compilation failed |
| `429 Too Many Requests` | Rate limit exceeded |
| `502 Bad Gateway` | Renderer service error |
| `503 Service Unavailable` | Renderer not configured |

## Request Limits

### PDF Rendering (/v1/renders-sync)
- Max body size: **1 MiB** (configurable via `SYNC_RENDER_MAX_BODY_BYTES`)
- Timeout: **8 seconds** (configurable via `SYNC_RENDER_TIMEOUT_SECONDS`)
- Max template size: 200KB
- Max compilation time: 20 seconds
- Max PDF size: 20MB

### Math Rendering (/v1/math-sync)
- Max body size: **64 KiB** (configurable via `MATH_SYNC_MAX_BODY_BYTES`)
- Max math input: 32KB
- Timeout: **8 seconds** (configurable via `MATH_SYNC_TIMEOUT_SECONDS`)

## Example Usage

### 0) Set your base URL and API key

```bash
export BASE_URL="https://latexlite.com"
export API_KEY="your-api-key-here"
```

### 1) Inline template without data (JSON body)

```bash
curl -sS -X POST "${BASE_URL}/v1/renders-sync" \
  -H "Authorization: Bearer ${API_KEY}" \
  -H "Content-Type: application/json" \
  -o output.pdf \
  -d '{
    "template": "\\documentclass{article}\n\\begin{document}\nHello, World!\n\\end{document}"
  }'
```

### 2) Inline template with data (JSON body)

```bash
curl -sS -X POST "${BASE_URL}/v1/renders-sync" \
  -H "Authorization: Bearer ${API_KEY}" \
  -H "Content-Type: application/json" \
  -o output.pdf \
  -d '{
    "template": "\\documentclass{article}\n\\begin{document}\nHello, [[.Who]]!\n\\end{document}",
    "data": { "Who": "world" }
  }'
```

### 3) Raw `.tex` file without data injection

Send a self-contained LaTeX file directly (no JSON escaping needed):

```bash
curl -sS -X POST "${BASE_URL}/v1/renders-sync" \
  -H "Authorization: Bearer ${API_KEY}" \
  -H "Content-Type: text/plain" \
  --data-binary @templates/simple.tex \
  -o simple-from-file.pdf
```

> **Note:** You can use `text/plain`, `text/x-tex`, or `application/x-tex` as the Content-Type. The file must be self-contained (no `[[.Field]]` placeholders) when using this method.

### 4) File upload with data injection (multipart)

When your template has `[[.Field]]` placeholders and you want to inject data:

```bash
curl -sS -X POST "${BASE_URL}/v1/renders-sync" \
  -H "Authorization: Bearer ${API_KEY}" \
  -F "template=@templates/invoice.tex;type=text/plain" \
  -F 'data={"CompanyName":"Acme Corp","InvoiceNumber":"INV-001","ClientName":"Client Ltd","Items":[{"Description":"Service","Qty":"1","UnitPrice":"$100","Total":"$100"}],"TotalDue":"$100"};type=application/json' \
  -o invoice-from-file.pdf
```

**Alternative:** Read data from a JSON file:

```bash
curl -sS -X POST "${BASE_URL}/v1/renders-sync" \
  -H "Authorization: Bearer ${API_KEY}" \
  -F "template=@templates/invoice.tex;type=text/plain" \
  -F "data=$(cat data/invoice.json);type=application/json" \
  -o invoice.pdf
```

Example `invoice.json`:

```json
{
  "CompanyName": "Acme Digital Ltd",
  "InvoiceNumber": "INV-2024-0042",
  "ClientName": "Widgets & Co Ltd",
  "Items": [
    { "Description": "Website Redesign", "Qty": "1", "UnitPrice": "£2,500.00", "Total": "£3,000.00" }
  ],
  "TotalDue": "£3,000.00"
}
```

### 5) Request JSON response (debugging)

Useful for inspecting the base64-encoded PDF:

```bash
curl -sS -X POST "${BASE_URL}/v1/renders-sync" \
  -H "Authorization: Bearer ${API_KEY}" \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -d '{
    "template": "\\documentclass{article}\n\\begin{document}\nHello, [[.Who]]!\n\\end{document}",
    "data": { "Who": "world" }
  }'
```

### 6) Math: render LaTeX equation to PNG

```bash
curl -sS -X POST "${BASE_URL}/v1/math-sync" \
  -H "Authorization: Bearer ${API_KEY}" \
  -H "Content-Type: application/json" \
  -o equation.png \
  -d '{
    "math": "$\\int_0^1 x^2 \\, dx = \\frac{1}{3}$"
  }'
```

> **Note:** In JSON, backslashes must be escaped. Use `\\int`, `\\frac`, `\\,` etc. However, `\n` (newline) and `\t` (tab) are JSON escape sequences and should remain single backslash.

### 7) Math: request JSON response

```bash
curl -sS -X POST "${BASE_URL}/v1/math-sync" \
  -H "Authorization: Bearer ${API_KEY}" \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -d '{
    "math": "$E = mc^2$"
  }'
```

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
- `"render timed out. Try optimizing your template."` (408)
- `"math must start and end with $ (or $$)"` (400)

## Best Practices

1. **Choose the right input type**:
   - Use **`application/json`** for inline templates with data injection
   - Use **`text/plain`** (or `text/x-tex`, `application/x-tex`) for raw `.tex` files without data - no JSON escaping needed
   - Use **`multipart/form-data`** when uploading `.tex` files that contain `[[.Field]]` placeholders and need data injection

2. **Escape LaTeX characters in JSON** - Backslashes must be doubled: `\` becomes `\\` for LaTeX commands like `\\int`, `\\frac`, `\\$`, etc. However, `\n` (newline) and `\t` (tab) are JSON escape sequences and should remain single backslash.
   ```json
   {
     "template": "\\documentclass{article}\n\\begin{document}\nHello \\textbf{World}!\n\\end{document}"
   }
   ```

3. **Handle rate limits** - Respect the `X-RateLimit-*` headers:
   - `X-RateLimit-Limit`: Maximum requests allowed per minute
   - `X-RateLimit-Remaining`: Requests remaining in current window
   - `X-RateLimit-Reset`: Unix timestamp when the limit resets

4. **Math input format** - Math strings must start and end with `$` or `$$`

5. **Error handling** - Always check `success` field and handle non-200 status codes

6. **Optimize templates** - Keep templates under 200KB and avoid computationally expensive LaTeX packages to stay within the 8-second timeout

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.