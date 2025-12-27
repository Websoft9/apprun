#!/bin/bash
# Auto-sync Sprint-Story-Module Mapping table in global README
# Scans all Story files and updates the global index

set -e

GLOBAL_README="docs/sprint-artifacts/README.md"

if [ ! -f "$GLOBAL_README" ]; then
    echo "❌ Global README not found: $GLOBAL_README"
    exit 1
fi

echo "🔄 Syncing Global Stories Index..."

# Generate new mapping table
{
    echo "| Sprint | Story & Description | Module | Status |"
    echo "|--------|---------------------|--------|--------|"
    
    for sprint_dir in $(find docs/sprint-artifacts -maxdepth 1 -type d -name "sprint-*" | sort); do
        sprint_name=$(basename "$sprint_dir")
        
        # Check if sprint has story files
        story_count=$(find "$sprint_dir" -maxdepth 1 -name "story-*.md" 2>/dev/null | wc -l)
        if [ "$story_count" -eq 0 ]; then
            continue
        fi
        
        for story_file in $(find "$sprint_dir" -maxdepth 1 -name "story-*.md" | sort); do
            # Extract metadata
            story_num=$(basename "$story_file" | grep -oP 'story-\K\d+')
            title=$(grep -m1 "^# Story" "$story_file" | sed 's/^# Story [0-9]*: //')
            module=$(grep -m1 "^\*\*Module\*\*:" "$story_file" | sed 's/.*: //' | xargs)
            status=$(grep -m1 "^\*\*Status\*\*:" "$story_file" | sed 's/.*: //' | xargs)
            
            # Format sprint name
            sprint_display=$(echo "$sprint_name" | sed 's/sprint-/Sprint-/')
            
            echo "| $sprint_display | Story $story_num: $title | $module | $status |"
        done
    done
} > /tmp/global-mapping.txt

# Check if markers exist
if ! grep -q "<!-- MAPPING_TABLE_START -->" "$GLOBAL_README"; then
    echo "⚠️  Warning: Markers not found in $GLOBAL_README"
    echo "    Please add <!-- MAPPING_TABLE_START --> and <!-- MAPPING_TABLE_END --> around the table"
    exit 1
fi

# Replace table in global README (between markers)
awk '
    /<!-- MAPPING_TABLE_START -->/ {
        print;
        system("cat /tmp/global-mapping.txt");
        skip=1;
        next;
    }
    /<!-- MAPPING_TABLE_END -->/ {
        skip=0;
    }
    !skip {
        print;
    }
' "$GLOBAL_README" > /tmp/global-readme-new.txt

mv /tmp/global-readme-new.txt "$GLOBAL_README"

echo "✅ Global index synced: $GLOBAL_README"
echo "📊 Total stories indexed: $(grep -c "^| Sprint-" /tmp/global-mapping.txt || echo 0)"
