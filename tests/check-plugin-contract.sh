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

python3 - "$repo_root/.agents/plugins/marketplace.json" "$plugin_root/plugin.json" "$plugin_root/.codex-plugin/plugin.json" "$repo_root" <<'PY'
import json
import sys
from pathlib import Path

marketplace_path, portable_path, compatibility_path, repo_root = sys.argv[1:]
with open(portable_path, encoding="utf-8") as handle:
    portable = json.load(handle)
with open(compatibility_path, encoding="utf-8") as handle:
    compatibility = json.load(handle)
with open(marketplace_path, encoding="utf-8") as handle:
    marketplace = json.load(handle)

schema = "https://agent-plugins.org/schemas/1.0.0/plugin.schema.json"
if portable.get("$schema") != schema:
    raise SystemExit("portable manifest has the wrong Agent Plugins schema")
required_portable_fields = (
    "$schema",
    "name",
    "version",
    "description",
    "author",
    "homepage",
    "repository",
    "license",
    "keywords",
    "extensions",
)
for field in required_portable_fields:
    if field not in portable:
        raise SystemExit(f"portable manifest field is missing: {field}")
if portable.get("name") != "llm-wiki":
    raise SystemExit("portable manifest name must be llm-wiki")
if not portable.get("version"):
    raise SystemExit("portable manifest version is missing")
if "skills" in portable:
    raise SystemExit("portable manifest must rely on root skills/ discovery")
if "interface" in portable:
    raise SystemExit("portable interface metadata must be under extensions.com.openai")

extensions = portable.get("extensions")
if not isinstance(extensions, dict) or not isinstance(extensions.get("com.openai"), dict):
    raise SystemExit("portable manifest extensions.com.openai is missing or invalid")
openai_extension = extensions["com.openai"]
portable_interface = openai_extension.get("interface", {})
if not isinstance(portable_interface, dict):
    raise SystemExit("portable manifest extensions.com.openai.interface is invalid")
for field in ("displayName", "shortDescription", "longDescription"):
    if not portable_interface.get(field):
        raise SystemExit(f"portable manifest interface.{field} is missing")

if compatibility.get("name") != portable.get("name"):
    raise SystemExit("portable and compatibility manifest names diverge")
if compatibility.get("version") != portable.get("version"):
    raise SystemExit("portable and compatibility manifest versions diverge")
for field in ("description", "author", "homepage", "repository", "license", "keywords"):
    if compatibility.get(field) != portable.get(field):
        raise SystemExit(f"portable and compatibility manifest {field} diverge")
if compatibility.get("skills") != "./skills/":
    raise SystemExit("compatibility manifest skills path must be ./skills/")
if compatibility.get("interface") != portable_interface:
    raise SystemExit("portable and compatibility interface metadata diverge")
if portable_interface.get("displayName") != "LLM Wiki":
    raise SystemExit("portable display name must be LLM Wiki")

plugin_root = Path(portable_path).parent
def check_asset(raw_path, label):
    if not isinstance(raw_path, str) or not raw_path.startswith("./"):
        raise SystemExit(f"{label} must be a plugin-relative ./ path")
    candidate = (plugin_root / raw_path).resolve()
    try:
        candidate.relative_to(plugin_root.resolve())
    except ValueError:
        raise SystemExit(f"{label} escapes the plugin root")
    if not candidate.is_file():
        raise SystemExit(f"{label} does not point to a file: {raw_path}")

check_asset(portable_interface.get("logo"), "portable interface.logo")
check_asset(compatibility.get("interface", {}).get("logo"), "compatibility interface.logo")

license_path = Path(repo_root) / "LICENSE"
if not license_path.is_file():
    raise SystemExit("root LICENSE is missing")
license_text = license_path.read_text(encoding="utf-8")
if "MIT License" not in license_text or "Santiago Migoni" not in license_text:
    raise SystemExit("root LICENSE is not the declared MIT license")

if marketplace.get("name") != "llm-wiki":
    raise SystemExit("marketplace name must be llm-wiki")
if marketplace.get("interface", {}).get("displayName") != "LLM Wiki":
    raise SystemExit("marketplace display name must be LLM Wiki")
entries = [entry for entry in marketplace.get("plugins", []) if isinstance(entry, dict)]
entry = next((entry for entry in entries if entry.get("name") == portable.get("name")), None)
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
    raise SystemExit("compatibility manifest path is missing")
if not (Path(repo_root) / "plugins" / "llm-wiki" / "plugin.json").is_file():
    raise SystemExit("portable manifest path is missing")
PY

required_paths=(
  ".agents/plugins/marketplace.json"
  "INSTALL.md"
  "README.md"
  "CHANGELOG.md"
  "plugins/llm-wiki/.codex-plugin/plugin.json"
  "plugins/llm-wiki/plugin.json"
  "plugins/llm-wiki/assets/logo.png"
  "LICENSE"
  "plugins/llm-wiki/docs/wiki-architecture.md"
  "plugins/llm-wiki/docs/provenance.md"
  "plugins/llm-wiki/docs/synthesis-format.md"
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
  "plugins/llm-wiki/scripts/llm-wiki"
  "plugins/llm-wiki/scripts/wiki-migrate-provenance"
  "plugins/llm-wiki/scripts/README.md"
  "plugins/llm-wiki/scripts/llm_wiki/frontmatter.py"
  "plugins/llm-wiki/scripts/llm_wiki/inventory.py"
  "plugins/llm-wiki/scripts/llm_wiki/hashes.py"
  "plugins/llm-wiki/scripts/llm_wiki/links.py"
  "plugins/llm-wiki/scripts/llm_wiki/validate.py"
  "plugins/llm-wiki/scripts/llm_wiki/provenance.py"
  "plugins/llm-wiki/scripts/tests/test_cli.py"
  "plugins/llm-wiki/scripts/tests/test_frontmatter.py"
  "plugins/llm-wiki/scripts/tests/test_provenance.py"
  "tests/fixtures/empty-vault/AGENTS.md"
  "tests/fixtures/empty-vault/wiki/index.md"
  "tests/fixtures/empty-vault/wiki/overview.md"
  "tests/fixtures/empty-vault/wiki/log.md"
)

for relative_path in "${required_paths[@]}"; do
  test -e "$repo_root/$relative_path" || fail "missing required path: $relative_path"
done
test -x "$plugin_root/scripts/wiki-migrate-provenance" || fail "provenance migration entrypoint is not executable"

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
rg -q "docs/provenance\\.md" "$plugin_root/references/vault-protocol.md" || fail "provenance policy is not linked"
rg -q "migrate-provenance" "$plugin_root/scripts/README.md" || fail "provenance migration is not documented"
rg -q "INV-EXTRACTION-MISSING" "$plugin_root/scripts/llm_wiki/inventory.py" || fail "inventory warnings are not implemented"
rg -q -- "--revert-to" "$plugin_root/scripts/llm_wiki/cli.py" || fail "hash reversion classification is not exposed"
rg -q "missing_headings" "$plugin_root/scripts/llm_wiki/links.py" || fail "heading resolution is not implemented"
rg -q "index_missing" "$plugin_root/scripts/llm_wiki/links.py" || fail "index route coverage is not implemented"
rg -q "raw/sources/<slug>" "$plugin_root/docs/wiki-architecture.md" || fail "architecture source path is missing"
rg -q "wiki/pages/" "$plugin_root/docs/wiki-architecture.md" || fail "architecture canonical page path is missing"
rg -q "scalability\\.md" "$repo_root/README.md" || fail "scale policy is not linked"
rg -q "assess-vault-scale\\.sh" "$repo_root/README.md" || fail "scale assessor is not documented"

PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover \
  -s "$plugin_root/scripts/tests" \
  -p 'test_*.py' >/dev/null || fail "deterministic Python tool tests failed"

validation_output="$(PYTHONDONTWRITEBYTECODE=1 python3 "$plugin_root/scripts/llm-wiki" validate "$repo_root/tests/fixtures/empty-vault")"
if [[ "$validation_output" != *"Status: VALID"* ]]; then
  fail "empty-vault deterministic validation failed"
fi

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
