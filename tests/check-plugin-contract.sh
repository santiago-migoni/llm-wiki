#!/usr/bin/env bash
set -euo pipefail

script_dir="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
repo_root="$(CDPATH= cd -- "$script_dir/.." && pwd)"

fail() {
  echo "FAIL: $*" >&2
  exit 1
}

command -v rg >/dev/null 2>&1 || fail "ripgrep is required"
command -v python3 >/dev/null 2>&1 || fail "python3 is required"

python3 - "$repo_root/.codex-plugin/plugin.json" <<'PY'
import json
import sys

path = sys.argv[1]
with open(path, encoding="utf-8") as handle:
    manifest = json.load(handle)

if manifest.get("name") != "llm-wiki":
    raise SystemExit("manifest name must be llm-wiki")
if not manifest.get("version"):
    raise SystemExit("manifest version is missing")
if manifest.get("skills") != "./skills/":
    raise SystemExit("manifest skills path must be ./skills/")
interface = manifest.get("interface", {})
for field in ("displayName", "shortDescription", "longDescription"):
    if not interface.get(field):
        raise SystemExit(f"manifest interface.{field} is missing")
PY

required_paths=(
  ".codex-plugin/plugin.json"
  "README.md"
  "CHANGELOG.md"
  "docs/wiki-architecture.md"
  "docs/plugin-roadmap.md"
  "docs/scalability.md"
  "docs/testing.md"
  "docs/host-compatibility.md"
  "references/vault-protocol.md"
  "skills/wiki-init/SKILL.md"
  "skills/wiki-init/assets/vault-template.md"
  "skills/wiki-ingest/SKILL.md"
  "skills/wiki-ingest/references/converting-documents.md"
  "skills/wiki-ingest/references/source-record.md"
  "skills/wiki-query/SKILL.md"
  "skills/wiki-query/references/context-modes.md"
  "skills/wiki-lint/SKILL.md"
  "skills/wiki-lint/references/lint-report.md"
  "tests/assess-vault-scale.sh"
  "tests/fixtures/empty-vault/AGENTS.md"
  "tests/fixtures/empty-vault/wiki/index.md"
  "tests/fixtures/empty-vault/wiki/overview.md"
  "tests/fixtures/empty-vault/wiki/log.md"
)

for relative_path in "${required_paths[@]}"; do
  test -e "$repo_root/$relative_path" || fail "missing required path: $relative_path"
done

required_dirs=(
  "tests/fixtures/empty-vault/raw/inbox"
  "tests/fixtures/empty-vault/raw/sources"
  "tests/fixtures/empty-vault/wiki/pages"
  "tests/fixtures/empty-vault/wiki/syntheses"
  "tests/fixtures/empty-vault/wiki/people"
  "tests/fixtures/empty-vault/wiki/concepts"
  "tests/fixtures/empty-vault/wiki/projects"
  "tests/fixtures/empty-vault/wiki/decisions"
)

for relative_dir in "${required_dirs[@]}"; do
  test -d "$repo_root/$relative_dir" || fail "missing required directory: $relative_dir"
done

for skill in wiki-init wiki-ingest wiki-query wiki-lint; do
  skill_path="$repo_root/skills/$skill/SKILL.md"
  rg -q "^name: $skill$" "$skill_path" || fail "invalid skill name: $skill"
  rg -q "^description: .+" "$skill_path" || fail "missing skill description: $skill"
done

for legacy_pattern in raw/archive raw/catalog revision-id document-id wiki/knowledge wiki/sources; do
  if rg -n "$legacy_pattern" \
      "$repo_root/README.md" \
      "$repo_root/skills/wiki-init" \
      "$repo_root/skills/wiki-ingest" \
      "$repo_root/skills/wiki-query"; then
    fail "obsolete operational reference found: $legacy_pattern"
  fi
done

rg -q "source-record\\.md" "$repo_root/skills/wiki-ingest/SKILL.md" || fail "ingest reference is not linked"
rg -q "context-modes\\.md" "$repo_root/skills/wiki-query/SKILL.md" || fail "query reference is not linked"
rg -q "lint-report\\.md" "$repo_root/skills/wiki-lint/SKILL.md" || fail "lint reference is not linked"
rg -q "docs/scalability\\.md" "$repo_root/skills/wiki-lint/SKILL.md" || fail "scale policy is not linked from wiki-lint"
rg -q "raw/sources/<slug>" "$repo_root/docs/wiki-architecture.md" || fail "architecture source path is missing"
rg -q "wiki/pages/" "$repo_root/docs/wiki-architecture.md" || fail "architecture canonical page path is missing"
rg -q "scalability\\.md" "$repo_root/README.md" || fail "scale policy is not linked"
rg -q "assess-vault-scale\\.sh" "$repo_root/README.md" || fail "scale assessor is not documented"

bash -n "$repo_root/tests/assess-vault-scale.sh" || fail "scale assessor has invalid Bash syntax"
scale_output="$(bash "$repo_root/tests/assess-vault-scale.sh" "$repo_root/tests/fixtures/empty-vault")"
if [[ "$scale_output" != *"Scale status: GREEN"* ]]; then
  fail "empty-vault scale assessment is not GREEN"
fi

scale_test_root="$(mktemp -d)"
trap 'rm -rf "$scale_test_root"' EXIT
cp -R "$repo_root/tests/fixtures/empty-vault/." "$scale_test_root/"
mv "$scale_test_root/AGENTS.md" "$scale_test_root/CLAUDE.md"

claude_scale_output="$(bash "$repo_root/tests/assess-vault-scale.sh" "$scale_test_root")"
if [[ "$claude_scale_output" != *"Schema: CLAUDE.md"* ]] || \
   [[ "$claude_scale_output" != *"Scale status: GREEN"* ]]; then
  fail "CLAUDE.md scale assessment is not supported"
fi

truncate -s 6M "$scale_test_root/wiki/syntheses/large.md"
large_scale_output="$(bash "$repo_root/tests/assess-vault-scale.sh" "$scale_test_root")"
if [[ "$large_scale_output" != *"Scale status: WATCH"* ]] || \
   [[ "$large_scale_output" != *"Largest synthesis Markdown file: 6.0 MiB"* ]]; then
  fail "large wiki Markdown is not included in scale assessment"
fi

cp "$scale_test_root/CLAUDE.md" "$scale_test_root/AGENTS.md"
if bash "$repo_root/tests/assess-vault-scale.sh" "$scale_test_root" > "$scale_test_root/both-schema.out" 2>&1; then
  fail "scale assessment accepts both schema files"
fi
rg -q "Multiple schema files: AGENTS.md, CLAUDE.md" "$scale_test_root/both-schema.out" || \
  fail "scale assessment does not report both schema files"

echo "plugin contract: ok"
