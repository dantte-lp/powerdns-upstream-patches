#!/usr/bin/env bash
# Verifies that every GitHub Action pinned by commit SHA actually resolves.
#
# This exists because a fabricated SHA is invisible until CI runs: the syntax is
# valid, review reads past it, and the failure lands on whoever next opens a
# pull request. It has already happened twice in this repository —
# astral-sh/setup-uv and go-task/setup-task — which is twice more than a rule
# without enforcement was worth.
#
# Also rejects floating tags. A tag is a mutable reference, and a workflow that
# holds a signing key is not the place for one.
#
# Usage:
#   scripts/check-action-pins.sh [file...]     defaults to .github/workflows/*.yml
#
# Requires `gh` authenticated. Skips with a warning when it is not available,
# so the hook does not block an offline commit — CI runs the same check.
set -euo pipefail

readonly SHA_RE='[0-9a-f]{40}'
readonly USES_RE='uses:[[:space:]]*([A-Za-z0-9._-]+/[A-Za-z0-9._-]+)@([^[:space:]]+)'

files=("$@")
if [ ${#files[@]} -eq 0 ]; then
  mapfile -t files < <(find .github/workflows -name '*.yml' -o -name '*.yaml' 2>/dev/null)
fi
[ ${#files[@]} -gt 0 ] || { echo "no workflow files found"; exit 0; }

if ! command -v gh >/dev/null 2>&1 || ! gh auth status >/dev/null 2>&1; then
  echo "check-action-pins: gh unavailable or unauthenticated — skipping" >&2
  echo "check-action-pins: CI runs the same check, so a bad pin still fails there" >&2
  exit 0
fi

declare -i verified=0 floating=0 missing=0
declare -A seen=()

for file in "${files[@]}"; do
  [ -f "$file" ] || continue
  while IFS= read -r line; do
    [[ "$line" =~ $USES_RE ]] || continue
    repo="${BASH_REMATCH[1]}"
    ref="${BASH_REMATCH[2]}"

    # Local and reusable-workflow references have no upstream to verify.
    [[ "$repo" == ./* ]] && continue

    if ! [[ "$ref" =~ ^$SHA_RE$ ]]; then
      printf 'FLOATING  %s@%s\n' "$repo" "$ref" >&2
      printf '          pin by 40-character commit SHA with a "# vX.Y.Z" comment\n' >&2
      floating+=1
      continue
    fi

    key="${repo}@${ref}"
    [ -n "${seen[$key]+x}" ] && continue
    seen[$key]=1

    if gh api "repos/${repo}/commits/${ref}" --jq '.sha' >/dev/null 2>&1; then
      printf 'ok        %-34s %s\n' "$repo" "${ref:0:12}"
      verified+=1
    else
      printf 'NOT FOUND %s@%s\n' "$repo" "$ref" >&2
      printf '          this SHA does not exist upstream — look it up, do not recall it:\n' >&2
      printf '          gh api repos/%s/git/ref/tags/<tag> --jq .object.sha\n' "$repo" >&2
      missing+=1
    fi
  done < "$file"
done

echo
if [ $(( floating + missing )) -gt 0 ]; then
  printf 'check-action-pins: %d verified, %d floating, %d not found\n' \
    "$verified" "$floating" "$missing" >&2
  exit 1
fi
printf 'check-action-pins: %d pins verified\n' "$verified"
