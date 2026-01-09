#!/usr/bin/env python3
"""
Generate Story Index Table from sprint-status.yaml

Outputs a formatted table showing: Sprint | Story ID | Title | Status | Epic
Can output to terminal or save as Markdown file
"""

import yaml
import sys
import argparse
from pathlib import Path
from datetime import datetime

def load_sprint_status(yaml_path):
    """Load sprint-status.yaml"""
    with open(yaml_path, 'r', encoding='utf-8') as f:
        return yaml.safe_load(f)

def generate_index_table(data, output_format='terminal'):
    """Generate story index table
    
    Args:
        data: Parsed YAML data
        output_format: 'terminal' or 'markdown'
    """
    
    # Build epic lookup
    epic_map = {epic['id']: epic['name'] for epic in data.get('epics', [])}
    
    # Collect all stories with their sprint info
    stories = []
    for story in data.get('stories', []):
        epic_name = epic_map.get(story.get('epic', ''), 'N/A')
        # Extract sprint from story ID (e.g., story-01 → sprint-0)
        story_id = story['id']
        sprint = f"sprint-{story_id.split('-')[1][0]}" if '-' in story_id else 'N/A'
        
        stories.append({
            'sprint': sprint,
            'id': story_id,
            'title': story['title'],
            'status': story['status'],
            'epic': epic_name,
            'epic_full': epic_name  # Keep full name for markdown
        })
    
    # Sort by sprint, then by story ID
    stories.sort(key=lambda x: (x['sprint'], x['id']))
    
    # Status icons
    status_icons = {
        'done': '✅',
        'in-progress': '🔄',
        'planning': '📝',
        'blocked': '🚫',
        'paused': '⏸️'
    }
    
    if output_format == 'markdown':
        return _generate_markdown(stories, status_icons)
    else:
        return _generate_terminal(stories, status_icons)


def _generate_terminal(stories, status_icons):
    """Generate terminal output"""
    lines = []
    lines.append("\n📋 Story Index Table")
    lines.append("=" * 120)
    lines.append(f"{'Sprint':<10} {'Story':<12} {'Title':<47} {'Status':<16} {'Epic':<30}")
    lines.append("-" * 120)
    
    current_sprint = None
    for story in stories:
        # Truncate for terminal display
        title = story['title']
        if len(title) > 45:
            title = title[:42] + "..."
        epic = story['epic']
        if len(epic) > 28:
            epic = epic[:25] + "..."
            
        # Add separator between sprints
        if current_sprint and current_sprint != story['sprint']:
            lines.append("-" * 120)
        current_sprint = story['sprint']
        
        icon = status_icons.get(story['status'], '❓')
        status_display = f"{icon} {story['status']}"
        
        row = f"{story['sprint']:<10} {story['id']:<12} {title:<47} {status_display:<16} {epic:<30}"
        lines.append(row)
    
    lines.append("=" * 120)
    
    # Add summary
    status_counts = {}
    for story in stories:
        status_counts[story['status']] = status_counts.get(story['status'], 0) + 1
    
    lines.append(f"\n📊 Summary: {len(stories)} stories")
    for status, count in sorted(status_counts.items()):
        icon = status_icons.get(status, '❓')
        lines.append(f"  {icon} {status.capitalize()}: {count}")
    lines.append("")
    
    return '\n'.join(lines)


def _generate_markdown(stories, status_icons):
    """Generate Markdown output"""
    lines = []
    
    # Header
    lines.append("# Story Index")
    lines.append("")
    lines.append(f"**Generated**: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    lines.append("")
    lines.append("---")
    lines.append("")
    
    # Summary
    status_counts = {}
    for story in stories:
        status_counts[story['status']] = status_counts.get(story['status'], 0) + 1
    
    lines.append("## Summary")
    lines.append("")
    lines.append(f"**Total Stories**: {len(stories)}")
    lines.append("")
    for status, count in sorted(status_counts.items()):
        icon = status_icons.get(status, '❓')
        lines.append(f"- {icon} **{status.capitalize()}**: {count}")
    lines.append("")
    lines.append("---")
    lines.append("")
    
    # Table by sprint
    current_sprint = None
    for story in stories:
        if current_sprint != story['sprint']:
            if current_sprint is not None:
                lines.append("")
            current_sprint = story['sprint']
            lines.append(f"## {current_sprint.title()}")
            lines.append("")
            lines.append("| Story ID | Title | Status | Epic |")
            lines.append("|----------|-------|--------|------|")
        
        icon = status_icons.get(story['status'], '❓')
        status_display = f"{icon} {story['status']}"
        
        # Escape pipe characters in content
        title = story['title'].replace('|', '\\|')
        epic = story['epic_full'].replace('|', '\\|')
        
        lines.append(f"| {story['id']} | {title} | {status_display} | {epic} |")
    
    lines.append("")
    lines.append("---")
    lines.append("")
    lines.append("*Generated from `sprint-status.yaml`*")
    lines.append("")
    
    return '\n'.join(lines)

def main():
    # Parse arguments
    parser = argparse.ArgumentParser(description='Generate story index table from sprint-status.yaml')
    parser.add_argument('--format', choices=['terminal', 'markdown'], default='terminal',
                       help='Output format (default: terminal)')
    parser.add_argument('--output', '-o', help='Output file path (only for markdown format)')
    args = parser.parse_args()
    
    # Find sprint-status.yaml
    yaml_path = Path(__file__).parent.parent / 'docs' / 'sprint-artifacts' / 'sprint-status.yaml'
    
    if not yaml_path.exists():
        print(f"❌ Error: {yaml_path} not found", file=sys.stderr)
        sys.exit(1)
    
    try:
        data = load_sprint_status(yaml_path)
        output = generate_index_table(data, args.format)
        
        if args.format == 'markdown':
            # Determine output path
            if args.output:
                output_path = Path(args.output)
            else:
                output_path = Path(__file__).parent.parent / 'docs' / 'sprint-artifacts' / 'story-index.md'
            
            # Write to file
            output_path.write_text(output, encoding='utf-8')
            print(f"✅ Story index generated: {output_path}")
        else:
            # Print to terminal
            print(output)
            
    except Exception as e:
        print(f"❌ Error generating story index: {e}", file=sys.stderr)
        import traceback
        traceback.print_exc()
        sys.exit(1)

if __name__ == '__main__':
    main()
