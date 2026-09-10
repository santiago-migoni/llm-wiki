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

plugin_root="$repo_root/plugins/llm-wiki"

python3 - "$repo_root/.agents/plugins/marketplace.json" "$plugin_root/.codex-plugin/plugin.json" "$repo_root" <<'PY'
import json
import sys
from pathlib import Path

marketplace_path, manifest_path, repo_root = sys.argv[1:]
with open(manifest_path, encoding="utf-8") as handle:
    manifest = json.load(handle)
with open(marketplace_path, encoding="utf-8") as handle:
    marketplace = json.load(handle)

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

if marketplace.get("name") != "llm-wiki":
    raise SystemExit("marketplace name must be llm-wiki")
entries = [entry for entry in marketplace.get("plugins", []) if isinstance(entry, dict)]
entry = next((entry for entry in entries if entry.get("name") == "llm-wiki"), None)
if entry is None:
    raise SystemExit("marketplace entry for llm-wiki is missing")
if entry.get("source", {}).get("path") != "./plugins/llm-wiki":
    raise SystemExit("marketplace source path must be ./plugins/llm-wiki")
if entry.get("policy") != {
    "installation": "AVAILABLE",
    "authentication": "ON_INSTALL",
}:
    raise SystemExit("marketplace policy is invalid")
if entry.get("category") != "Productivity":
    raise SystemExit("marketplace category is invalid")
if not (Path(repo_root) / "plugins" / "llm-wiki" / ".codex-plugin" / "plugin.json").is_file():
    raise SystemExit("plugin manifest path is missing")
PY

required_paths=(
  ".agents/plugins/marketplace.json"
  "INSTALL.md"
  "README.md"
  "CHANGELOG.md"
  "plugins/llm-wiki/.codex-plugin/plugin.json"
  "plugins/llm-wiki/docs/wiki-architecture.md"
  "plugins/llm-wiki/docs/plugin-roadmap.md"
  "plugins/llm-wiki/docs/scalability.md"
  "plugins/llm-wiki/docs/testing.md"
  "plugins/llm-wiki/docs/host-compatibility.md"
  "plugins/llm-wiki/references/vault-protocol.md"
  "plugins/llm-wiki/skills/wiki-init/SKILL.md"
  "plugins/llm-wiki/skills/wiki-init/assets/vault-template.md"
  "plugins/llm-wiki/skills/wiki-ingest/SKILL.md"
  "plugins/llm-wiki/skills/wiki-ingest/references/converting-documents.md"
  "plugins/llm-wiki/skills/wiki-ingest/references/source-record.md"
  "plugins/llm-wiki/skills/wiki-query/SKILL.md"
  "plugins/llm-wiki/skills/wiki-query/references/context-modes.md"
  "plugins/llm-wiki/skills/wiki-lint/SKILL.md"
  "plugins/llm-wiki/skills/wiki-lint/references/lint-report.md"
  "plugins/llm-wiki/tests/assess-vault-scale.sh"
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
  skill_path="$plugin_root/skills/$skill/SKILL.md"
  rg -q "^name: $skill$" "$skill_path" || fail "invalid skill name: $skill"
  rg -q "^description: .+" "$skill_path" || fail "missing skill description: $skill"
done

for legacy_pattern in raw/archive raw/catalog revision-id document-id wiki/knowledge wiki/sources; do
  if rg -n "$legacy_pattern" \
      "$repo_root/README.md" \
      "$plugin_root/skills/wiki-init" \
      "$plugin_root/skills/wiki-ingest" \
      "$plugin_root/skills/wiki-query"; then
    fail "obsolete operational reference found: $legacy_pattern"
  fi
done

rg -q "source-record\\.md" "$plugin_root/skills/wiki-ingest/SKILL.md" || fail "ingest reference is not linked"
rg -q "context-modes\\.md" "$plugin_root/skills/wiki-query/SKILL.md" || fail "query reference is not linked"
rg -q "lint-report\\.md" "$plugin_root/skills/wiki-lint/SKILL.md" || fail "lint reference is not linked"
rg -q "docs/scalability\\.md" "$plugin_root/skills/wiki-lint/SKILL.md" || fail "scale policy is not linked from wiki-lint"
rg -q "raw/sources/<slug>" "$plugin_root/docs/wiki-architecture.md" || fail "architecture source path is missing"
rg -q "wiki/pages/" "$plugin_root/docs/wiki-architecture.md" || fail "architecture canonical page path is missing"
rg -q "scalability\\.md" "$repo_root/README.md" || fail "scale policy is not linked"
rg -q "assess-vault-scale\\.sh" "$repo_root/README.md" || fail "scale assessor is not documented"

bash -n "$plugin_root/tests/assess-vault-scale.sh" || fail "scale assessor has invalid Bash syntax"
scale_output="$(bash "$plugin_root/tests/assess-vault-scale.sh" "$repo_root/tests/fixtures/empty-vault")"
if [[ "$scale_output" != *"Scale status: GREEN"* ]]; then
  fail "empty-vault scale assessment is not GREEN"
fi

scale_test_root="$(mktemp -d)"
trap 'rm -rf "$scale_test_root"' EXIT
cp -R "$repo_root/tests/fixtures/empty-vault/." "$scale_test_root/"
mv "$scale_test_root/AGENTS.md" "$scale_test_root/CLAUDE.md"

claude_scale_output="$(bash "$plugin_root/tests/assess-vault-scale.sh" "$scale_test_root")"
if [[ "$claude_scale_output" != *"Schema: CLAUDE.md"* ]] || \
   [[ "$claude_scale_output" != *"Scale status: GREEN"* ]]; then
  fail "CLAUDE.md scale assessment is not supported"
fi

truncate -s 6M "$scale_test_root/wiki/syntheses/large.md"
large_scale_output="$(bash "$plugin_root/tests/assess-vault-scale.sh" "$scale_test_root")"
if [[ "$large_scale_output" != *"Scale status: WATCH"* ]] || \
   [[ "$large_scale_output" != *"Largest synthesis Markdown file: 6.0 MiB"* ]]; then
  fail "large wiki Markdown is not included in scale assessment"
fi

cp "$scale_test_root/CLAUDE.md" "$scale_test_root/AGENTS.md"
if bash "$plugin_root/tests/assess-vault-scale.sh" "$scale_test_root" > "$scale_test_root/both-schema.out" 2>&1; then
  fail "scale assessment accepts both schema files"
fi
rg -q "Multiple schema files: AGENTS.md, CLAUDE.md" "$scale_test_root/both-schema.out" || \
  fail "scale assessment does not report both schema files"

echo "plugin contract: ok"
