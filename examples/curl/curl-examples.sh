#!/bin/bash
set -euo pipefail

# API_KEY="<your-api-key>"
# URL="https://latexlite.com"

echo "LaTeX API Examples"
echo "=================="
echo ""

# 1) Inline template without data (JSON body)
echo "1. Inline template without data -> output.pdf"
curl -sS -X POST "${URL}/v1/renders-sync" \
  -H "Authorization: Bearer ${API_KEY}" \
  -H "Content-Type: application/json" \
  -o output.pdf \
  -d '{
    "template": "\\documentclass{article}\n\\begin{document}\nHello, World!\n\\end{document}"
  }'
echo "   Downloaded output.pdf"
echo ""

# 2) Inline template with data (JSON body)
echo "2. Inline template with data -> output-with-data.pdf"
curl -sS -X POST "${URL}/v1/renders-sync" \
  -H "Authorization: Bearer ${API_KEY}" \
  -H "Content-Type: application/json" \
  -o output-with-data.pdf \
  -d '{
    "template": "\\documentclass{article}\n\\begin{document}\nHello, [[.Who]]!\n\\end{document}",
    "data": { "Who": "world" }
  }'
echo "   Downloaded output-with-data.pdf"
echo ""

# 3) Raw .tex file without data injection
echo "3. Raw .tex file without data injection -> simple-from-file.pdf"
curl -sS -X POST "${URL}/v1/renders-sync" \
  -H "Authorization: Bearer ${API_KEY}" \
  -H "Content-Type: text/plain" \
  --data-binary @../../templates/simple.tex \
  -o simple-from-file.pdf
echo "   Downloaded simple-from-file.pdf"
echo ""

# 4) File upload with data injection (multipart) - inline JSON
echo "4. File upload with data injection (inline JSON) -> invoice-inline.pdf"
curl -sS -X POST "${URL}/v1/renders-sync" \
  -H "Authorization: Bearer ${API_KEY}" \
  -F "template=@../../templates/invoice.tex;type=text/plain" \
  -F 'data={"CompanyName":"Acme Corp","InvoiceNumber":"INV-001","ClientName":"Client Ltd","Items":[{"Description":"Service","Qty":"1","UnitPrice":"$100","Total":"$100"}],"TotalDue":"$100"};type=application/json' \
  -o invoice-inline.pdf
echo "   Downloaded invoice-inline.pdf"
echo ""

# 5) File upload with data injection (multipart) - from JSON file
echo "5. File upload with data injection (from JSON file) -> invoice.pdf"
curl -sS -X POST "${URL}/v1/renders-sync" \
  -H "Authorization: Bearer ${API_KEY}" \
  -F "template=@../../templates/invoice.tex;type=text/plain" \
  -F "data=$(cat ../../templates/invoice.json);type=application/json" \
  -o invoice.pdf
echo "   Downloaded invoice.pdf"
echo ""

# # 6) Request JSON response (debugging)
# echo "6. Request JSON response (debugging)"
# curl -sS -X POST "${URL}/v1/renders-sync" \
#   -H "Authorization: Bearer ${API_KEY}" \
#   -H "Accept: application/json" \
#   -H "Content-Type: application/json" \
#   -d '{
#     "template": "\\documentclass{article}\n\\begin{document}\nHello, [[.Who]]!\n\\end{document}",
#     "data": { "Who": "world" }
#   }' | jq '.'
# echo ""

# # 7) Math: render LaTeX equation to PNG
# echo "7. Math: render LaTeX equation to PNG -> equation.png"
# curl -sS -X POST "${URL}/v1/math-sync" \
#   -H "Authorization: Bearer ${API_KEY}" \
#   -H "Content-Type: application/json" \
#   -o equation.png \
#   -d '{
#     "math": "$\\int_0^1 x^2 \\, dx = \\frac{1}{3}$"
#   }'
# echo "   Downloaded equation.png"
# echo ""

# # 8) Math: request JSON response
# echo "8. Math: request JSON response"
# curl -sS -X POST "${URL}/v1/math-sync" \
#   -H "Authorization: Bearer ${API_KEY}" \
#   -H "Accept: application/json" \
#   -H "Content-Type: application/json" \
#   -d '{
#     "math": "$E = mc^2$"
#   }' | jq '.'
# echo ""

# # 9) Check health
# echo "9. Checking API health"
# curl -sS -H "Authorization: Bearer ${API_KEY}" \
#   "${URL}/health" | jq '.'
# echo ""

# echo "Examples complete!"