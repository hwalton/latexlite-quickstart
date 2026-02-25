#!/bin/bash
set -euo pipefail

API_KEY="<your-api-key>"
BASE_URL="https://latexlite.com"

echo "🚀 LaTeX API Examples"

# 1) Render (no polling) — writes PDF directly
echo "1. Render (renders-sync) -> sync.pdf ..."
curl -sS -X POST "${BASE_URL}/v1/renders-sync" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -o sync.pdf \
  -d '{
    "template": "\\documentclass{article}\n\\begin{document}\nHello, [[.Who]]!\n\\end{document}",
    "data": { "Who": "sync world" }
  }'

echo "Downloaded sync.pdf"

# 1b) (Optional) Render returning JSON (base64 PDF)
echo "1b. Render (renders-sync) JSON response ..."
curl -sS -X POST "${BASE_URL}/v1/renders-sync" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -d '{
    "template": "\\documentclass{article}\n\\begin{document}\nHello, [[.Who]]!\n\\end{document}",
    "data": { "Who": "json world" }
  }' | jq '.'

# 2) Sync math -> PNG
echo "4. Sync math (math-sync) -> equation.png ..."
curl -sS -X POST "${BASE_URL}/v1/math-sync" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -o equation.png \
  -d '{
    "math": "$\\int_0^1 x^2 \\, dx = \\frac{1}{3}$"
  }'
echo "Downloaded equation.png"

# 2b) Sync math -> JSON response
echo "4b. Sync math (math-sync) JSON response ..."
curl -sS -X POST "${BASE_URL}/v1/math-sync" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -d '{
    "math": "$E = mc^2$"
  }' | jq '.'

# 3) Check health
echo "5. Checking API health..."
curl -sS -H "Authorization: Bearer $API_KEY" \
  "${BASE_URL}/health" | jq '.'

echo "Examples complete!"