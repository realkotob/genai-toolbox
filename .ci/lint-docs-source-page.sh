#!/bin/bash
set -e

python3 - << 'EOF'
import os
import re
import sys
from pathlib import Path

integration_dir = Path("./docs/en/integrations")
if not integration_dir.exists():
    print("Info: Directory './docs/en/integrations' not found. Skipping linting.")
    sys.exit(0)

ALLOWED_ORDER = [
    "About",
    "Available Tools",
    "Requirements",
    "Example",
    "Reference",
    "Advanced Usage",
    "Troubleshooting",
    "Additional Resources"
]
REQUIRED = {"About", "Example", "Reference"}

# Regex to catch any variation of the list-tools shortcode, including parameters
SHORTCODE_PATTERN = r"\{\{<\s*list-tools.*?>\}\}"

has_errors = False

# Find all _index.md files inside the subdirectories of integrations/
for filepath in integration_dir.rglob("_index.md"):
    # Skip the top-level integrations/_index.md if it exists
    if filepath.parent == integration_dir:
        continue

    with open(filepath, "r", encoding="utf-8") as f:
        content = f.read()

    # Separate YAML frontmatter from the markdown body
    match = re.match(r'^\s*---\s*\n(.*?)\n---\s*(.*)', content, re.DOTALL)
    if match:
        frontmatter = match.group(1)
        body = match.group(2)
    else:
        frontmatter = ""
        body = content

    # If the file has no markdown content (metadata placeholder only), skip it entirely
    if not body.strip():
        continue

    file_errors = False

    # 1. Check Frontmatter Title
    title_source = frontmatter if frontmatter else content
    title_match = re.search(r"^title:\s*[\"']?(.*?)[\"']?\s*$", title_source, re.MULTILINE)
    if not title_match or not title_match.group(1).strip().endswith("Source"):
        found_title = title_match.group(1) if title_match else "None"
        print(f"[{filepath}] Error: Frontmatter title must end with 'Source'. Found: '{found_title}'")
        file_errors = True

    # 2. Check Shortcode Placement ONLY IF "Available Tools" heading is present
    tools_section_match = re.search(r"^##\s+Available Tools\s*(.*?)(?=^##\s|\Z)", body, re.MULTILINE | re.DOTALL)
    if tools_section_match:
        if not re.search(SHORTCODE_PATTERN, tools_section_match.group(1)):
            print(f"[{filepath}] Error: The list-tools shortcode must be placed under the '## Available Tools' heading.")
            file_errors = True
    else:
        # Prevent edge case where shortcode is used but the heading was forgotten
        if re.search(SHORTCODE_PATTERN, body):
            print(f"[{filepath}] Error: A list-tools shortcode was found, but the '## Available Tools' heading is missing.")
            file_errors = True

    # 3. Strip code blocks from body to avoid linting example markdown headings
    clean_body = re.sub(r"```.*?```", "", body, flags=re.DOTALL)

    # 4. Check H1 Headings
    if re.search(r"^#\s+\w+", clean_body, re.MULTILINE):
        print(f"[{filepath}] Error: H1 headings (#) are forbidden in the body.")
        file_errors = True

    # 5. Check H2 Headings
    h2s = re.findall(r"^##\s+(.*)", clean_body, re.MULTILINE)
    h2s = [h2.strip() for h2 in h2s]

    # Missing Required
    missing = REQUIRED - set(h2s)
    if missing:
        print(f"[{filepath}] Error: Missing required H2 headings: {missing}")
        file_errors = True

    # Unauthorized Headings
    unauthorized = set(h2s) - set(ALLOWED_ORDER)
    if unauthorized:
        print(f"[{filepath}] Error: Unauthorized H2 headings found: {unauthorized}")
        file_errors = True

    # Strict Ordering
    filtered_h2s = [h for h in h2s if h in ALLOWED_ORDER]
    expected_order = [h for h in ALLOWED_ORDER if h in h2s]
    if filtered_h2s != expected_order:
        print(f"[{filepath}] Error: Headings are out of order.")
        print(f"  Expected: {expected_order}")
        print(f"  Found:    {filtered_h2s}")
        file_errors = True

    if file_errors:
        has_errors = True

if has_errors:
    print("Linting failed. Please fix the structure errors above.")
    sys.exit(1)
else:
    print("Success: All Source pages passed structure validation.")
EOF