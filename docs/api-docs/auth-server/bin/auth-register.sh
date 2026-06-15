#!/bin/bash

# -e: exit immediately on any command failure
# -u: treat unset environment variables as an error
# -o pipefail: return the exit code of the first failed command in a pipeline
set -euo pipefail

RID=$(cat /proc/sys/kernel/random/uuid)
echo "Generated Request ID: ${RID}"

data="{\"host\": \"${BK_AUTHSERVER_HOST}\"}"

res=$(curl -s -X POST \
  -H "Content-Type: application/json" \
  -H "X-Bkapi-User-Name: ${BK_AUTH_USER_NAME}" \
  -H "X-Bkapi-App-Code: ${BK_APP_CODE}" \
  -H "X-Bkapi-Request-Id: ${RID}" \
  -H "X-Bk-Tenant-Id: ${BK_TENANT_ID}" \
  --data "${data}" \
  "http://${BK_AUTHSERVER_ENDPOINT}/api/v1/auth/init/authcenter"
)

echo "${res}"
if ! echo "${res}" | grep -qE '"code"[[:space:]]*:[[:space:]]*"?0"?'; then
  echo "auth center migration failed."
  exit 1
fi
