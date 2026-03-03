#!/usr/bin/env bash
set -euo pipefail

BIN="${1:-./supashift}"
if [[ ! -x "$BIN" ]]; then
  echo "Binary no ejecutable: $BIN" >&2
  echo "Compila primero: go build -o supashift ./cmd/supashift" >&2
  exit 1
fi

ROOT_TMP="$(mktemp -d)"
trap 'rm -rf "$ROOT_TMP"' EXIT

export HOME="$ROOT_TMP/home"
export XDG_CONFIG_HOME="$ROOT_TMP/xdg"
export SUPASHIFT_PASSPHRASE="smoke-passphrase"
mkdir -p "$HOME/.supabase"
echo "GLOBAL_SENTINEL_TOKEN" > "$HOME/.supabase/access-token"

echo "[1/8] init"
"$BIN" init >/dev/null

echo "[2/8] crear 3 perfiles"
"$BIN" profile add client-a --account "Client A" --token "tok_a" --tags cliente,prod --aliases a-prod >/dev/null
"$BIN" profile add client-b --account "Client B" --token "tok_b" --tags cliente,dev --aliases b-dev >/dev/null
"$BIN" profile add client-c --account "Client C" --token "tok_c" --tags cliente,stg --aliases c-stg >/dev/null

count="$("$BIN" profile ls | tail -n +2 | wc -l | tr -d ' ')"
[[ "$count" == "3" ]] || { echo "Se esperaban 3 perfiles, hay $count" >&2; exit 1; }

echo "[3/8] validar run aislado por perfil"
"$BIN" run client-a -- sh -c '[ "$SUPABASE_ACCESS_TOKEN" = "tok_a" ]'
"$BIN" run client-b -- sh -c '[ "$SUPABASE_ACCESS_TOKEN" = "tok_b" ]'
"$BIN" run client-c -- sh -c '[ "$SUPABASE_ACCESS_TOKEN" = "tok_c" ]'

echo "[4/8] validar no modifica token global supabase"
[[ "$(cat "$HOME/.supabase/access-token")" == "GLOBAL_SENTINEL_TOKEN" ]] || { echo "Token global fue modificado" >&2; exit 1; }

echo "[5/8] validar use/unuse snippet"
use_out="$("$BIN" use client-a)"
unuse_out="$("$BIN" unuse)"
echo "$use_out" | grep -q 'SUPABASE_ACCESS_TOKEN' || { echo "Snippet use inválido" >&2; exit 1; }
echo "$unuse_out" | grep -q 'unset SUPABASE_ACCESS_TOKEN' || { echo "Snippet unuse inválido" >&2; exit 1; }

echo "[6/8] validar migrate-from-supabase-cli desde archivo legacy"
echo "tok_legacy" > "$HOME/.supabase/access-token"
"$BIN" migrate-from-supabase-cli legacy --source file --account "Legacy" >/dev/null
"$BIN" run legacy -- sh -c '[ "$SUPABASE_ACCESS_TOKEN" = "tok_legacy" ]'

echo "[7/8] doctor"
"$BIN" doctor >/dev/null

echo "[8/8] tmux simultáneo (opcional si tmux existe)"
if command -v tmux >/dev/null 2>&1; then
  printf '1,2,3\n' | "$BIN" tmux --many >/dev/null
  tmux ls | grep -q 'supa-client-a' || { echo "Falta sesión tmux client-a" >&2; exit 1; }
  tmux ls | grep -q 'supa-client-b' || { echo "Falta sesión tmux client-b" >&2; exit 1; }
  tmux ls | grep -q 'supa-client-c' || { echo "Falta sesión tmux client-c" >&2; exit 1; }
else
  echo "tmux no instalado; se omite validación de sesiones"
fi

echo "OK: smoke test completado"
