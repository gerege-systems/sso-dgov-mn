#!/usr/bin/env bash
# Government SSO (sso.dgov.mn) — register a relying-party (OAuth2/OIDC client).
#
# Registers a confidential RP with BOTH a login redirect_uri AND a matching
# post_logout_redirect_uri, so RP-initiated logout does NOT fail with
# "post_logout_redirect_uri is not whitelisted". The RP libraries derive the
# post-logout URL as scheme://host/ from their redirect_uri — this script sets
# exactly that, so login and logout both work out of the box.
#
# NOTE: The in-app Applications admin UI/API already does this automatically
# (see internal/provider/adminapi: postLogoutFromRedirects). Prefer that UI.
# Use this script only for quick CLI registration on the server.
#
# Usage:
#   ./scripts/register-rp.sh "<App name>" https://app.dgov.mn [extra-callback-path]
# Example:
#   ./scripts/register-rp.sh "Government Template Platform" https://template.dgov.mn
set -euo pipefail

NAME="${1:?usage: register-rp.sh <name> <base_url e.g. https://app.dgov.mn> [callback_path]}"
BASE="${2:?base_url required, e.g. https://app.dgov.mn}"
CB_PATH="${3:-/sso/callback}"
BASE="${BASE%/}"                       # strip trailing slash
HYDRA_CONTAINER="${HYDRA_CONTAINER:-sso-dgov-mn-hydra-1}"
HYDRA_ADMIN="${HYDRA_ADMIN:-http://127.0.0.1:4445}"   # admin API inside the container

docker exec "$HYDRA_CONTAINER" hydra create oauth2-client \
  --endpoint "$HYDRA_ADMIN" \
  --name "$NAME" \
  --grant-type authorization_code,refresh_token \
  --response-type code \
  --scope openid,profile,email \
  --redirect-uri "${BASE}${CB_PATH}" \
  --post-logout-callback "${BASE}/" \
  --token-endpoint-auth-method client_secret_basic \
  --format json

echo >&2
echo ">> Registered RP '${NAME}'" >&2
echo ">>   redirect_uri            = ${BASE}${CB_PATH}" >&2
echo ">>   post_logout_redirect_uri= ${BASE}/" >&2
echo ">> Copy client_id / client_secret above into the RP's SSO_CLIENT_ID / SSO_CLIENT_SECRET." >&2
