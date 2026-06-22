#!/usr/bin/env bash
# claude-compliance-check.sh — Validate Claude Code structured response
# Usage: claude-compliance-check.sh /path/to/claude-response.json

set -euo pipefail

if [ $# -ne 1 ]; then
  echo "Usage: $0 /path/to/claude-response.json"
  exit 1
fi

RESPONSE_FILE="$1"

if [ ! -f "$RESPONSE_FILE" ]; then
  echo "ERROR: response file not found: $RESPONSE_FILE"
  exit 1
fi

SCHEMA="/home/ubuntu/projects/vexa/00-meta/claude-response-schema.json"
if [ ! -f "$SCHEMA" ]; then
  echo "ERROR: schema file not found: $SCHEMA"
  exit 1
fi

# Check if python and jsonschema available
if ! command -v python3 &> /dev/null; then
  echo "ERROR: python3 not found"
  exit 1
fi

python3 - <<EOF
import json, sys

schema_path = "${SCHEMA}"
response_path = "${RESPONSE_FILE}"

try:
    import jsonschema
except ImportError:
    print("ERROR: python jsonschema module not installed")
    sys.exit(1)

with open(schema_path, "r") as f:
    schema = json.load(f)

with open(response_path, "r") as f:
    data = json.load(f)

# Extract the actual result if wrapped in Claude Code wrapper
if isinstance(data, dict) and "result" in data and isinstance(data["result"], str):
    # Try to parse the result string as JSON
    try:
        inner = json.loads(data["result"])
        if isinstance(inner, dict):
            data = inner
    except json.JSONDecodeError:
        # result is free text; fail
        print("ERROR: response 'result' field is not valid JSON structured data")
        sys.exit(1)

try:
    jsonschema.validate(instance=data, schema=schema)
except jsonschema.ValidationError as e:
    print(f"ERROR: schema validation failed: {e.message}")
    print(f"Path: {list(e.path)}")
    sys.exit(1)

# Check all 10 compliance confirmations are true
confirmation = data["compliance"]["confirmation"]
if not all(confirmation):
    missing = [i+1 for i, v in enumerate(confirmation) if not v]
    print(f"ERROR: compliance confirmation missing at points: {missing}")
    sys.exit(1)

# Check scope constraints
scope = data["scope"]
if not scope.get("only_02_application"):
    print("ERROR: scope.only_02_application must be true")
    sys.exit(1)
if not scope.get("no_push"):
    print("ERROR: scope.no_push must be true")
    sys.exit(1)
if not scope.get("no_root_edit"):
    print("ERROR: scope.no_root_edit must be true")
    sys.exit(1)

print("OK: Claude Code response is compliant ✅")
print(f"Context files read: {len(data['context']['files_read'])}")
print(f"Skills used: {', '.join(data['skills']['used'])}")
print(f"Files changed: {len(data['work']['files_changed'])}")
print(f"Lines of impact: {data['work']['lines_of_impact']}")
print(f"Verification gates passed: {data['verification']['gates_passed']}")
EOF
