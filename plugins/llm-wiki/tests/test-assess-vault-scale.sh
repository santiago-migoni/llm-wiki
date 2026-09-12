#!/usr/bin/env bash
set -euo pipefail

fail() {
  echo "FAIL: $*" >&2
  exit 1
}

plugin_root="$(cd "$(dirname "$0")/.." && pwd)"
repo_root="$(cd "$plugin_root/../.." && pwd)"
assessor="$plugin_root/tests/assess-vault-scale.sh"
empty_fixture="$repo_root/tests/fixtures/empty-vault"
test_root="$(mktemp -d)"
last_status=0
last_output=""

cleanup() {
  chmod -R u+rwX "$test_root" 2>/dev/null || true
  rm -rf "$test_root"
}
trap cleanup EXIT

run_assessor() {
  local root="$1"
  if last_output="$(bash "$assessor" "$root" 2>&1)"; then
    last_status=0
  else
    last_status=$?
  fi
}

new_case() {
  local case_root
  case_root="$(mktemp -d "$test_root/case.XXXXXX")"
  cp -R "$empty_fixture/." "$case_root/"
  printf '%s\n' "$case_root"
}

assert_status() {
  local root="$1"
  local expected="$2"
  run_assessor "$root"
  [[ "$last_status" -eq 0 ]] || fail "expected successful assessment for $root: $last_output"
  [[ "$last_output" == *"Scale status: $expected"* ]] || \
    fail "expected Scale status: $expected, got: $last_output"
}

assert_invalid_type() {
  local root="$1"
  local path="$2"
  run_assessor "$root"
  [[ "$last_status" -eq 1 ]] || fail "expected invalid path type for $path: $last_output"
  [[ "$last_output" == *"Invalid path type: $path must be a directory"* ]] || \
    fail "missing stable invalid-type message for $path: $last_output"
}

assert_invalid_schema_type() {
  local root="$1"
  run_assessor "$root"
  [[ "$last_status" -eq 1 ]] || fail "expected invalid schema type: $last_output"
  [[ "$last_output" == *"Invalid path type: AGENTS.md must be a regular file"* ]] || \
    fail "missing stable invalid-schema message: $last_output"
}

assert_status "$empty_fixture" "GREEN"
assert_status "$repo_root/tests/fixtures/populated-vault" "GREEN"
assert_status "$repo_root/tests/fixtures/legacy-vault" "GREEN"

case_root="$(new_case)"
for slug in one/shared two/shared; do
  record="$case_root/raw/sources/$slug"
  mkdir -p "$record/assets"
  touch "$record/source.txt" "$record/extracted.md"
done
assert_status "$case_root" "GREEN"
[[ "$last_output" == *"Current source records: 2"* ]] || \
  fail "nested source records were not counted as leaves"

run_assessor "$repo_root/tests/fixtures/malformed-vault"
[[ "$last_status" -eq 1 ]] || fail "malformed fixture was accepted: $last_output"
[[ "$last_output" == *"Scale status: INVALID"* ]] || fail "malformed fixture has no INVALID status"
[[ "$last_output" == *"Invalid source slugs: Bad_Slug"* ]] || fail "invalid slug was not reported"
[[ "$last_output" == *"Sources missing extracted.md: 1"* ]] || fail "missing extraction was not reported"
[[ "$last_output" == *"Unexpected source entries:"* ]] || fail "unexpected source entry was not reported"

run_assessor "$repo_root/tests/fixtures/duplicate-source-files"
[[ "$last_status" -eq 1 ]] || fail "duplicate-source fixture was accepted: $last_output"
[[ "$last_output" == *"Sources with multiple current originals: 1"* ]] || \
  fail "multiple current originals were not reported"
[[ "$last_output" == *"Multiple current originals: policy (source.docx, source.pdf)"* ]] || \
  fail "multiple source filenames were not reported"

for invalid_path in raw/inbox raw/sources wiki/pages; do
  case_root="$(new_case)"
  path="$case_root/$invalid_path"
  backup="$path-directory"
  mv "$path" "$backup"
  touch "$path"
  assert_invalid_type "$case_root" "$invalid_path"
done

case_root="$(new_case)"
mv "$case_root/AGENTS.md" "$case_root/AGENTS.original"
mkdir "$case_root/AGENTS.md"
assert_invalid_schema_type "$case_root"

case_root="$(new_case)"
mkdir "$case_root/wiki/pages/inaccessible"
chmod 000 "$case_root/wiki/pages/inaccessible"
run_assessor "$case_root"
chmod 700 "$case_root/wiki/pages/inaccessible"
[[ "$last_status" -eq 1 ]] || fail "inaccessible directory was accepted: $last_output"
[[ "$last_output" == *"Inaccessible directory: wiki/pages/inaccessible"* ]] || \
  fail "inaccessible directory was not reported: $last_output"

for count in 79 80 100 101; do
  case_root="$(new_case)"
  for ((index = 1; index <= count; index++)); do
    record="$case_root/raw/sources/source-$index"
    mkdir -p "$record"
    touch "$record/source.txt" "$record/extracted.md"
  done
  if [[ "$count" -lt 80 ]]; then
    expected="GREEN"
  elif [[ "$count" -le 100 ]]; then
    expected="WATCH"
  else
    expected="DERIVED-SEARCH-CANDIDATE"
  fi
  assert_status "$case_root" "$expected"
done

for count in 199 200 300 301; do
  case_root="$(new_case)"
  for ((index = 1; index <= count; index++)); do
    touch "$case_root/wiki/pages/page-$index.md"
  done
  if [[ "$count" -lt 200 ]]; then
    expected="GREEN"
  elif [[ "$count" -le 300 ]]; then
    expected="WATCH"
  else
    expected="DERIVED-SEARCH-CANDIDATE"
  fi
  assert_status "$case_root" "$expected"
done

echo "scale assessment: ok"
