#!/bin/bash

INPUT_FILE="emails.txt"
VALID_OUTPUT="valid_emails.txt"
INVALID_OUTPUT="invalid_emails.txt"

# Clear previous results
> "$VALID_OUTPUT"
> "$INVALID_OUTPUT"

while IFS= read -r email; do
  if [[ -z "$email" ]]; then
    continue
  fi

  # Make the curl request and parse the is_reachable field
  response=$(curl -s -X POST http://localhost:9100/v0/check_email \
    -H "Content-Type: application/json" \
    -d "{\"to_email\": \"$email\"}")

  # Extract is_reachable value
  reachable=$(echo "$response" | jq -r '.is_reachable')

  # Output formatted result
  echo "$email: $reachable"

  # Save to corresponding file
  if [[ "$reachable" == "valid" ]]; then
    echo "$email" >> "$VALID_OUTPUT"
  else
    echo "$email" >> "$INVALID_OUTPUT"
  fi

done < "$INPUT_FILE"
