# LaTeX Lite API - Quickstart Guide

Generate professional PDFs from LaTeX templates with a simple REST API.

## 🚀 Quick Start

### 1. Get API Access
```bash
# Demo API key (rate limited)
export LATEX_API_KEY="demo-key-1234567890abcdef"
export LATEX_API_URL="https://your-api-domain.com"

```

### 2. First PDF in 30 Seconds

```bash
curl -X POST $LATEX_API_URL/v1/renders \
  -H "Authorization: Bearer $LATEX_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "template": "\\documentclass{article}\\begin{document}Hello {{.Name}}!\\end{document}",
    "data": {"Name": "World"}
  }' \
  | jq '.data.id'

# Get your PDF (use job ID from above)
curl -H "Authorization: Bearer $LATEX_API_KEY" \
  $LATEX_API_URL/v1/renders/job_123456789/pdf \
  -o hello.pdf
```

## 📚 Examples

| Language   | Description                    | Run                          |
|------------|--------------------------------|------------------------------|
| **Go**     | Full client with error handling| `cd examples/go && go run .`|
| **Curl**   | Shell scripts for testing     | `cd examples/curl && ./examples.sh` |
| **Python** | Coming soon                    | -                            |
| **Node.js**| Coming soon                    | -                            |

## 🧾 Templates

### Invoice Template
```bash
# Use the pre-built invoice template
cat templates/invoice.tex | jq -Rs . | \
curl -X POST $LATEX_API_URL/v1/renders \
  -H "Authorization: Bearer $LATEX_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "template": .,
    "data": {
      "InvoiceNumber": "INV-001",
      "Date": "December 14, 2025",
      "CustomerName": "Acme Corp",
      "CustomerAddress": "123 Main St\\nNew York, NY 10001",
      "Items": [
        {"Description": "Web Development", "Amount": "2500.00"},
        {"Description": "Hosting Setup", "Amount": "500.00"}
      ],
      "Total": "3000.00"
    }
  }'
```

### Business Letter
See `templates/letter.tex` for a professional letter template.

## 🔑 API Reference

### Authentication
Include your API key in every request:
```
Authorization: Bearer your-api-key-here
```

### Rate Limits
- Demo key: 60 requests/minute
- Custom key: 100 requests/minute
- Check headers: `X-RateLimit-Remaining`

### Endpoints

#### `POST /v1/renders`
Create a render job.

**Request:**
```json
{
  "template": "LaTeX template with {{.Variables}}",
  "data": {"Variables": "values"}
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "job_123456789",
    "status": "queued",
    "created_at": "2025-01-15T10:30:00Z",
    "expires_at": "2025-01-16T10:30:00Z"
  }
}
```

#### `GET /v1/renders/{id}`
Check job status.

#### `GET /v1/renders/{id}/pdf`
Download PDF (when status = "succeeded").

#### `GET /health`
API health check (no auth required).

## ❌ Error Handling

| Status | Meaning |
|--------|---------|
| `401` | Invalid API key |
| `429` | Rate limit exceeded |
| `400` | Invalid template or data |
| `409` | PDF not ready (still processing) |

## 🎯 Best Practices

1. **Handle async processing** - Jobs are queued and processed in background
2. **Escape LaTeX characters** - `\` becomes `\\` in JSON strings  
3. **Poll for completion** - Check status every 2-5 seconds
4. **Cache PDFs** - Jobs expire after 24 hours
5. **Handle rate limits** - Respect the `X-RateLimit-*` headers

## 🤝 Contributing

Found a bug or want to add an example?

1. Fork this repository
2. Create your feature branch (`git checkout -b feature/python-example`)
3. Commit your changes (`git commit -am 'Add Python client example'`)
4. Push to the branch (`git push origin feature/python-example`)
5. Create a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- 🐛 **Issues**: [GitHub Issues](https://github.com/yourusername/latexlite-quickstart/issues)
- 📖 **Documentation**: [API Docs](https://your-api-domain.com/docs)
- 💬 **Community**: [Discussions](https://github.com/yourusername/latexlite-quickstart/discussions)