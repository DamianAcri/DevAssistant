#!/bin/bash
SECRET="my-super-secret"

PAYLOAD="test_payload.json"

URL="http://localhost:8080/webhook/github"

SIG=$(printf '%s' "$(cat $PAYLOAD)" | openssl dgst -sha256 -hmac "$SECRET" -hex | sed 's/^.* //')

echo "Using signature: sha256=$SIG"
echo ""

curl -i -X POST "$URL" \
  -H "Content-Type: application/json" \
  -H "X-GitHub-Event: pull_request" \
  -H "X-Hub-Signature-256: sha256=$SIG" \
  --data-binary @"$PAYLOAD"