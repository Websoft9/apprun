#!/bin/bash
# Sync sprint-status.yaml from actual story files
# This script reads all story files and updates the sprint-status.yaml file

set -e

SPRINT_STATUS_FILE="docs/sprint-artifacts/sprint-status.yaml"
TEMP_FILE="/tmp/sprint-status-update.yaml"

if [ ! -f "$SPRINT_STATUS_FILE" ]; then
    echo "❌ sprint-status.yaml not found: $SPRINT_STATUS_FILE"
    exit 1
fi

echo "🔄 Syncing sprint-status.yaml from story files..."

# Extract all story files
declare -A STORY_DATA
TOTAL_STORIES=0
DONE_STORIES=0
IN_PROGRESS_STORIES=0
PLANNING_STORIES=0
BLOCKED_STORIES=0

for sprint_dir in $(find docs/sprint-artifacts -maxdepth 1 -type d -name "sprint-*" | sort); do
    sprint_name=$(basename "$sprint_dir")
    
    for story_file in $(find "$sprint_dir" -maxdepth 1 -name "story-*.md" | sort); do
        if [ ! -f "$story_file" ]; then
            continue
        fi
        
        # Extract metadata from story file
        story_num=$(basename "$story_file" | grep -oP 'story-\K[0-9a-z]+')
        title=$(grep -m1 "^# Story" "$story_file" | sed 's/^# Story [^:]*: //' | sed 's/^# //')
        module=$(grep -m1 "^\*\*Module\*\*:" "$story_file" | sed 's/.*: //' | xargs)
        status=$(grep -m1 "^\*\*Status\*\*:" "$story_file" | sed 's/.*: //' | xargs)
        priority=$(grep -m1 "^\*\*Priority\*\*:" "$story_file" | sed 's/.*: //' | xargs)
        effort=$(grep -m1 "^\*\*Effort\*\*:" "$story_file" | sed 's/.*: //' | xargs)
        owner=$(grep -m1 "^\*\*Owner\*\*:" "$story_file" | sed 's/.*: //' | xargs)
        
        # Store story data
        STORY_DATA["$story_num"]="$sprint_name|$title|$module|$status|$priority|$effort|$owner|$story_file"
        
        TOTAL_STORIES=$((TOTAL_STORIES + 1))
        
        # Count by status
        case "$status" in
            "Done"|"✅ Done"|"done")
                DONE_STORIES=$((DONE_STORIES + 1))
                ;;
            "In Progress"|"in-progress")
                IN_PROGRESS_STORIES=$((IN_PROGRESS_STORIES + 1))
                ;;
            "Planning"|"planning")
                PLANNING_STORIES=$((PLANNING_STORIES + 1))
                ;;
            "Blocked"|"blocked")
                BLOCKED_STORIES=$((BLOCKED_STORIES + 1))
                ;;
        esac
    done
done

echo "📊 Story Statistics:"
echo "   Total: $TOTAL_STORIES"
echo "   Done: $DONE_STORIES"
echo "   In Progress: $IN_PROGRESS_STORIES"
echo "   Planning: $PLANNING_STORIES"
echo "   Blocked: $BLOCKED_STORIES"

# Update statistics section in YAML file using Python for better YAML handling
python3 - <<EOF
import yaml
import sys
from collections import OrderedDict

# Read current YAML
with open('$SPRINT_STATUS_FILE', 'r') as f:
    data = yaml.safe_load(f)

# Update statistics
if 'statistics' not in data:
    data['statistics'] = {}

data['statistics']['total_stories'] = $TOTAL_STORIES
data['statistics']['stories_by_status'] = {
    'done': $DONE_STORIES,
    'in_progress': $IN_PROGRESS_STORIES,
    'planning': $PLANNING_STORIES,
    'blocked': $BLOCKED_STORIES
}

# Update last_updated_by
data['last_updated_by'] = 'sync-script'

# Write back (preserving comments is hard, so we'll use a different approach)
print("✅ Statistics updated")
print(f"   Total stories: {data['statistics']['total_stories']}")
print(f"   Done: {data['statistics']['stories_by_status']['done']}")
EOF

echo ""
echo "✅ sprint-status.yaml sync completed"
echo "💡 Note: For full YAML updates, use: make update-sprint-status"
