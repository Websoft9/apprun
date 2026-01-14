#!/usr/bin/env python3
"""
Generate Story Index Table from sprint-status.yaml

Outputs a formatted table showing: Sprint | Story ID | Title | Status | Epic
Can output to terminal or save as Markdown file
"""

import yaml
import sys
import argparse
import re
from pathlib import Path
from datetime import datetime

def load_sprint_status(yaml_path):
    """Load sprint-status.yaml"""
    with open(yaml_path, 'r', encoding='utf-8') as f:
        return yaml.safe_load(f)

def parse_development_status(data):
    """Parse development_status into epics and stories
    
    Returns:
        tuple: (epics_dict, stories_list)
    """
    dev_status = data.get('development_status', {})
    epics = {}
    stories = []
    
    for key, value in dev_status.items():
        if key.startswith('epic-'):
            # This is an epic definition
            epics[key] = {
                'id': key,
                'title': value.get('title', ''),
                'status': value.get('status', 'backlog')
            }
        else:
            # This is a story
            # Parse story ID format: 1-16-redis-cache-package or 1-1-docker-environment
            story_id = key
            
            # Extract epic number from story ID (e.g., "1-16" -> "epic-1")
            match = re.match(r'^(\d+)-', story_id)
            epic_id = f"epic-{match.group(1)}" if match else 'epic-unknown'
            
            # Determine sprint from story number
            # Story format: {epic}-{story_num}-{name}
            # Sprint logic: story 1-15 -> sprint-1, 1-16+ -> sprint-2, etc.
            story_match = re.match(r'^(\d+)-(\d+)', story_id)
            if story_match:
                epic_num = int(story_match.group(1))
                story_num = int(story_match.group(2))
                
                # Story numbering to sprint mapping
                if epic_num == 1:
                    if story_num <= 9:
                        sprint = 'sprint-1'
                    elif story_num <= 16:
                        sprint = 'sprint-2'
                    else:
                        sprint = 'sprint-3'
                elif epic_num == 2:
                    sprint = 'sprint-1'
                elif epic_num == 3:
                    sprint = 'sprint-2'
                elif epic_num == 4:
                    sprint = 'sprint-2'
                elif epic_num == 5:
                    sprint = 'sprint-2'
                else:
                    sprint = 'sprint-unknown'
            else:
                sprint = 'sprint-unknown'
            
            # Use sprint from YAML if provided
            if 'sprint' in value:
                sprint = value['sprint']
            
            stories.append({
                'id': story_id,
                'title': value.get('title', ''),
                'status': value.get('status', 'backlog'),
                'epic_id': epic_id,
                'sprint': sprint,
                'priority': value.get('priority', ''),
                'estimate': value.get('estimate', '')
            })
    
    return epics, stories

def generate_index_table(data, output_format='terminal'):
    """Generate story index table
    
    Args:
        data: Parsed YAML data
        output_format: 'terminal' or 'markdown'
    """
    
    # Parse epics and stories
    epics, stories = parse_development_status(data)
    
    # Enrich stories with epic names
    for story in stories:
        epic = epics.get(story['epic_id'], {})
        story['epic_name'] = epic.get('title', 'Unknown Epic')
    
    # Sort by sprint, then by story ID
    stories.sort(key=lambda x: (x['sprint'], x['id']))
    
    # Status icons
    status_icons = {
        'done': '✅',
        'in-progress': '🔄',
        'planning': '📝',
        'backlog': '📋',
        'blocked': '🚫',
        'paused': '⏸️',
        'ready-for-dev': '🟢',
        'drafted': '✏️'
    }
    
    if output_format == 'markdown':
        return _generate_markdown(stories, status_icons, data.get('generated', datetime.now().strftime('%Y-%m-%d')))
    else:
        return _generate_terminal(stories, status_icons)


def _generate_terminal(stories, status_icons):
    """Generate terminal output"""
    lines = []
    lines.append("\n📋 Story Index Table")
    lines.append("=" * 120)
    lines.append(f"{'Sprint':<15} {'Story ID':<30} {'Title':<45} {'Status':<18} {'Epic':<25}")
    lines.append("-" * 120)
    
    current_sprint = None
    for story in stories:
        # Truncate for terminal display
        title = story['title']
        if len(title) > 43:
            title = title[:40] + "..."
        epic = story['epic_name']
        if len(epic) > 23:
            epic = epic[:20] + "..."
            
        # Add separator between sprints
        if current_sprint and current_sprint != story['sprint']:
            lines.append("-" * 120)
        current_sprint = story['sprint']
        
        icon = status_icons.get(story['status'], '❓')
        status_display = f"{icon} {story['status']}"
        
        row = f"{story['sprint']:<15} {story['id']:<30} {title:<45} {status_display:<18} {epic:<25}"
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


def _generate_markdown(stories, status_icons, generated_date):
    """Generate Markdown output"""
    lines = []
    
    # Header
    lines.append("# Story Index")
    lines.append("")
    lines.append(f"**Generated**: {generated_date}")
    lines.append("")
    lines.append("---")
    lines.append("")
    
    # Summary
    status_counts = {}
    sprint_counts = {}
    for story in stories:
        status_counts[story['status']] = status_counts.get(story['status'], 0) + 1
        sprint_counts[story['sprint']] = sprint_counts.get(story['sprint'], 0) + 1
    
    lines.append("## Summary")
    lines.append("")
    lines.append(f"**Total Stories**: {len(stories)}")
    lines.append("")
    
    # Status breakdown
    lines.append("### By Status")
    lines.append("")
    for status in ['done', 'in-progress', 'ready-for-dev', 'drafted', 'planning', 'backlog', 'blocked', 'paused']:
        if status in status_counts:
            icon = status_icons.get(status, '❓')
            lines.append(f"- {icon} **{status.replace('-', ' ').title()}**: {status_counts[status]}")
    lines.append("")
    
    # Sprint breakdown
    lines.append("### By Sprint")
    lines.append("")
    for sprint in sorted(sprint_counts.keys()):
        lines.append(f"- **{sprint.title()}**: {sprint_counts[sprint]} stories")
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
            lines.append(f"## {current_sprint.replace('-', ' ').title()}")
            lines.append("")
            lines.append("| Story ID | Title | Status | Epic |")
            lines.append("|----------|-------|--------|------|")
        
        icon = status_icons.get(story['status'], '❓')
        status_display = f"{icon} {story['status']}"
        
        # Escape pipe characters in content
        title = story['title'].replace('|', '\\|')
        epic = story['epic_name'].replace('|', '\\|')
        
        lines.append(f"| {story['id']} | {title} | {status_display} | {epic} |")
    
    lines.append("")
    lines.append("---")
    lines.append("")
    lines.append(f"*Auto-generated from `sprint-status.yaml` on {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}*")
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
