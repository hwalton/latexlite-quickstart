#!/bin/bash
set -euo pipefail

API_KEY="<your-api-key>"
BASE_URL="https://latexlite.com"

echo "🚀 LaTeX API Examples"

# 1) Sync render (no polling) — writes PDF directly
echo "1. Sync render (renders-sync) -> sync.pdf ..."
curl -sS -X POST "${BASE_URL}/v1/renders-sync" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -o sync.pdf \
  -d '{
    "template": "\\documentclass{article}\n\\begin{document}\nHello, [[.Who]]!\n\\end{document}",
    "data": { "Who": "sync world" }
  }'

echo "Downloaded sync.pdf"

# (Optional) Sync render returning JSON (base64 PDF)
echo "1b. Sync render (renders-sync) JSON response ..."
curl -sS -X POST "${BASE_URL}/v1/renders-sync" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -d '{
    "template": "\\documentclass{article}\n\\begin{document}\nHello, [[.Who]]!\n\\end{document}",
    "data": { "Who": "json world" }
  }' | jq '.'

# 2) Async: Simple document
echo "2. Creating simple document (async)..."
RESPONSE=$(curl -sS -X POST "${BASE_URL}/v1/renders" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "template": "\\documentclass{article}\\begin{document}Hello [[.Name]]!\\end{document}",
    "data": {"Name": "API User"}
  }')

echo "$RESPONSE" | jq '.'
JOB_ID=$(echo "$RESPONSE" | jq -r '.data.id')

# Wait and download
echo "2. Waiting for completion..."
sleep 2

curl -sS -H "Authorization: Bearer $API_KEY" \
  "${BASE_URL}/v1/renders/$JOB_ID/pdf" \
  -o simple.pdf

echo "Downloaded simple.pdf"

# 3) Async: Business letter
echo "3. Creating business letter (async)..."
LETTER_RESPONSE=$(curl -sS -X POST "${BASE_URL}/v1/renders" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
  "template": "\\documentclass{article}\\begin{document}\\noindent From: [[.YourName]]\\\\[[.YourAddress]]\\vspace{2em} To: [[.RecipientName]]\\\\[[.RecipientAddress]]\\vspace{2em} Dear [[.RecipientName]], [[.Message]]\\vspace{2em} Sincerely,\\\\[[.YourName]]\\end{document}",
  "data": {
    "YourName": "John Smith",
    "YourAddress": "123 Business St\\\\New York, NY 10001",
    "RecipientName": "Jane Doe",
    "RecipientAddress": "556 Client Ave\\\\Boston, MA 02101",
    "Message": "Thank you for your interest in our LaTeX API service. This letter demonstrates our template capabilities."
  }
}')

echo "$LETTER_RESPONSE" | jq '.'
LETTER_ID=$(echo "$LETTER_RESPONSE" | jq -r '.data.id')

echo "3. Waiting for completion..."
sleep 2

curl -sS -H "Authorization: Bearer $API_KEY" \
  "${BASE_URL}/v1/renders/$LETTER_ID/pdf" \
  -o letter.pdf

echo "Downloaded letter.pdf"

# 4) Sync math -> PNG
echo "4. Sync math (math-sync) -> equation.png ..."
curl -sS -X POST "${BASE_URL}/v1/math-sync" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -o equation.png \
  -d '{
    "math": "$\\int_0^1 x^2 \\, dx = \\frac{1}{3}$"
  }'
echo "Downloaded equation.png"

# 4b) Sync math -> JSON response
echo "4b. Sync math (math-sync) JSON response ..."
curl -sS -X POST "${BASE_URL}/v1/math-sync" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -d '{
    "math": "$E = mc^2$"
  }' | jq '.'

# 5) Check health
echo "5. Checking API health..."
curl -sS -H "Authorization: Bearer $API_KEY" \
  "${BASE_URL}/health" | jq '.'

echo "Examples complete!"