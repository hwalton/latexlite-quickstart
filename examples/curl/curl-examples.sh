#!/bin/bash

API_KEY="demo-key-1234567890abcdef"
BASE_URL="https://latexlite.com"

echo "🚀 LaTeX API Examples"

# Simple document
echo "1. Creating simple document..."
RESPONSE=$(curl -s -X POST $BASE_URL/v1/renders \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "template": "\\documentclass{article}\\begin{document}Hello [[.Name]]!\\end{document}",
    "data": {"Name": "API User"}
  }')

echo $RESPONSE | jq '.'
JOB_ID=$(echo $RESPONSE | jq -r '.data.id')

# Wait and download
echo "2. Waiting for completion..."
sleep 2

curl -H "Authorization: Bearer $API_KEY" \
  $BASE_URL/v1/renders/$JOB_ID/pdf \
  -o simple.pdf

echo "Downloaded simple.pdf"

# Business letter
echo "3. Creating business letter..."
LETTER_RESPONSE=$(curl -s -X POST $BASE_URL/v1/renders \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
  "template": "\\documentclass{article}\\begin{document}\\noindent From: [[.YourName]]\\\\[[.YourAddress]]\\vspace{2em} To: [[.RecipientName]]\\\\[[.RecipientAddress]]\\vspace{2em} Dear [[.RecipientName]], [[.Message]]\\vspace{2em} Sincerely,\\\\[[.YourName]]\\end{document}",
  "data": {
    "YourName": "John Smith",
    "YourAddress": "123 Business St\\\\New York, NY 10001",
    "RecipientName": "Jane Doe",
    "RecipientAddress": "456 Client Ave\\\\Boston, MA 02101",
    "Message": "Thank you for your interest in our LaTeX API service. This letter demonstrates our template capabilities."
  }
}')


echo $LETTER_RESPONSE | jq '.'
LETTER_ID=$(echo $LETTER_RESPONSE | jq -r '.data.id')

# Wait for the letter to be ready
echo "4. Waiting for completion..."
sleep 2

# Download the letter PDF
curl -H "Authorization: Bearer $API_KEY" \
  $BASE_URL/v1/renders/$LETTER_ID/pdf \
  -o letter.pdf

echo "Downloaded letter.pdf"

# Check health
echo "5. Checking API health..."
curl -s -H "Authorization: Bearer $API_KEY" \
  $BASE_URL/health | jq '.'

echo "Examples complete!"