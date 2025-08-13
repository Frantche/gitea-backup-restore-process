#!/usr/bin/env bash
set -euo pipefail

# --- Usage ---
# ./gitea_bootstrap.sh https://gitea.example.com alice 'MyS3cret!' my-repo "Titre de l'issue" "Corps de l'issue"
#
# Paramètres:
#   1: GITEA_URL       (ex: https://gitea.example.com)
#   2: GITEA_USER      (ex: alice)
#   3: GITEA_PASSWORD  (ex: MyS3cret!)
#   4: REPO_NAME       (ex: my-repo)
#   5: ISSUE_TITLE     (ex: "Bug: truc marche pas")
#   6: ISSUE_BODY      (ex: "Étapes pour reproduire ...")
#
# Dépendances: curl, jq

if ! command -v curl >/dev/null 2>&1; then
  echo "Erreur: curl est requis." >&2; exit 1
fi
if ! command -v jq >/dev/null 2>&1; then
  echo "Erreur: jq est requis." >&2; exit 1
fi

if [ "$#" -lt 5 ]; then
  echo "Usage: $0 <gitea_url> <user> <password> <repo_name> <issue_title> [issue_body]" >&2
  exit 1
fi

GITEA_URL="${1%/}"     # supprime un éventuel / final
GITEA_USER="$2"
GITEA_PASSWORD="$3"
REPO_NAME="$4"
ISSUE_TITLE="$5"
ISSUE_BODY="${6:-"Issue créée automatiquement."}"

# --- Fonction utilitaire pour afficher joliment les erreurs API ---
api_error() {
  local body="$1"
  # Essaie d'extraire 'message' ou 'error'
  if echo "$body" | jq -e . >/dev/null 2>&1; then
    echo "$(echo "$body" | jq -r '.message // .error // .err // . | tostring')" >&2
  else
    echo "$body" >&2
  fi
}

# --- 1) Créer le dépôt ---
echo "Création du dépôt '${REPO_NAME}' pour l'utilisateur ${GITEA_USER}…"
CREATE_REPO_PAYLOAD="$(jq -n --arg name "$REPO_NAME" '{
  name: $name,
  private: false,
  auto_init: true,     # crée une branche initiale avec README
  default_branch: "main"
}')"

# POST /api/v1/user/repos
create_repo_resp="$(curl -sS -w '\n%{http_code}' \
  -u "$GITEA_USER:$GITEA_PASSWORD" \
  -H 'Content-Type: application/json' \
  -X POST "$GITEA_URL/api/v1/user/repos" \
  -d "$CREATE_REPO_PAYLOAD")"

create_repo_body="$(echo "$create_repo_resp" | sed -n '1,$-1p' | sed '$d')"
create_repo_code="$(echo "$create_repo_resp" | tail -n1)"

if [ "$create_repo_code" -ge 300 ]; then
  echo "Échec de création du dépôt (HTTP $create_repo_code):"
  api_error "$create_repo_body"
  exit 1
fi

# Récupère owner et nom normalisé depuis la réponse
OWNER="$(echo "$create_repo_body" | jq -r '.owner.login // .owner.username // empty')"
REPO="$(echo "$create_repo_body" | jq -r '.name')"

if [ -z "${OWNER:-}" ] || [ -z "${REPO:-}" ]; then
  echo "Impossible de déterminer owner/repo depuis la réponse API:" >&2
  echo "$create_repo_body" | jq . >&2 || echo "$create_repo_body" >&2
  exit 1
fi

REPO_HTTP_URL="$(echo "$create_repo_body" | jq -r '.html_url // .ssh_url // .clone_url // empty')"
echo "Dépôt créé: ${OWNER}/${REPO}"
[ -n "$REPO_HTTP_URL" ] && echo "URL: $REPO_HTTP_URL"

# --- 2) Créer l'issue ---
echo "Création de l'issue dans ${OWNER}/${REPO}…"
CREATE_ISSUE_PAYLOAD="$(jq -n --arg title "$ISSUE_TITLE" --arg body "$ISSUE_BODY" '{
  title: $title,
  body: $body
}')"

# POST /api/v1/repos/{owner}/{repo}/issues
create_issue_resp="$(curl -sS -w '\n%{http_code}' \
  -u "$GITEA_USER:$GITEA_PASSWORD" \
  -H 'Content-Type: application/json' \
  -X POST "$GITEA_URL/api/v1/repos/'$OWNER'/'$REPO'/issues" \
  -d "$CREATE_ISSUE_PAYLOAD")"

create_issue_body="$(echo "$create_issue_resp" | sed -n '1,$-1p' | sed '$d')"
create_issue_code="$(echo "$create_issue_resp" | tail -n1)"

if [ "$create_issue_code" -ge 300 ]; then
  echo "Échec de création de l'issue (HTTP $create_issue_code):"
  api_error "$create_issue_body"
  exit 1
fi

ISSUE_URL="$(echo "$create_issue_body" | jq -r '.html_url // .url // empty')"
ISSUE_NUMBER="$(echo "$create_issue_body" | jq -r '.number // empty')"

echo "Issue créée: #${ISSUE_NUMBER:-?}"
[ -n "$ISSUE_URL" ] && echo "URL: $ISSUE_URL"

echo "Terminé ✅"