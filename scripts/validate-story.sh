#!/bin/bash
# Story document validation script
# Validates Story documents against metadata specification

set -e

STORY_FILE="$1"

if [ -z "$STORY_FILE" ]; then
    echo "Usage: $0 <story-file.md>"
    exit 1
fi

if [ ! -f "$STORY_FILE" ]; then
    echo "❌ File not found: $STORY_FILE"
    exit 1
fi

echo "🔍 Validating: $(basename "$STORY_FILE")"

# Required fields
REQUIRED_FIELDS=(
    "Priority"
    "Effort"
    "Owner"
    "Dependencies"
    "Status"
    "Module"
    "Issue"
)

# Check required fields
for field in "${REQUIRED_FIELDS[@]}"; do
    if ! grep -q "^\*\*${field}\*\*:" "$STORY_FILE"; then
        echo "❌ Missing required field: $field"
        exit 1
    fi
done

# Check priority format
PRIORITY=$(grep "^\*\*Priority\*\*:" "$STORY_FILE" | sed 's/.*: *//' | sed 's/ .*//' | tr -d ' ')
if [[ ! "$PRIORITY" =~ ^P[0-2]$ ]]; then
    echo "❌ Invalid priority: '$PRIORITY' (must be P0, P1, or P2)"
    exit 1
fi

# Check status format
STATUS=$(grep "^\*\*Status\*\*:" "$STORY_FILE" | sed 's/.*: *//' | xargs)
VALID_STATUSES=("Planning" "In Progress" "Done" "Blocked")
if [[ ! " ${VALID_STATUSES[@]} " =~ " ${STATUS} " ]]; then
    echo "❌ Invalid status: '$STATUS' (must be: Planning, In Progress, Done, or Blocked)"
    exit 1
fi

# Check module format
MODULE=$(grep "^\*\*Module\*\*:" "$STORY_FILE" | sed 's/.*: *//' | xargs)
VALID_MODULES=("Infrastructure" "Auth" "Storage" "Functions" "Management")
if [[ ! " ${VALID_MODULES[@]} " =~ " ${MODULE} " ]]; then
    echo "⚠️  Warning: Module '$MODULE' is not in standard list"
fi

# Check required sections
REQUIRED_SECTIONS=(
    "User Story"
    "Acceptance Criteria"
    "Implementation Tasks"
    "Test Cases"
    "Related Docs"
)

for section in "${REQUIRED_SECTIONS[@]}"; do
    if ! grep -q "^## ${section}" "$STORY_FILE"; then
        echo "❌ Missing required section: ## $section"
        exit 1
    fi
done

echo "✅ Story validation passed: $(basename "$STORY_FILE")"
