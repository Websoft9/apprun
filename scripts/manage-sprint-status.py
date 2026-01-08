#!/usr/bin/env python3
"""
Sprint Status YAML Manager (Simplified Version)
This tool helps BMad agents read and update sprint-status.yaml
Adapted for simplified structure with only epics and stories sections
"""

import yaml
import argparse
import sys
from pathlib import Path
from typing import Dict, List, Optional
from datetime import datetime

SPRINT_STATUS_FILE = Path("docs/sprint-artifacts/sprint-status.yaml")


class SprintStatusManager:
    def __init__(self, file_path: Path = SPRINT_STATUS_FILE):
        self.file_path = file_path
        self.data = self._load()

    def _load(self) -> Dict:
        """Load YAML file"""
        if not self.file_path.exists():
            raise FileNotFoundError(f"File not found: {self.file_path}")
        
        with open(self.file_path, 'r', encoding='utf-8') as f:
            return yaml.safe_load(f)

    def _save(self):
        """Save YAML file"""
        with open(self.file_path, 'w', encoding='utf-8') as f:
            yaml.dump(self.data, f, allow_unicode=True, default_flow_style=False, sort_keys=False)

    def get_epic(self, epic_id: str) -> Optional[Dict]:
        """Get epic by ID"""
        for epic in self.data.get('epics', []):
            if epic.get('id') == epic_id:
                return epic
        return None

    def get_story(self, story_id: str) -> Optional[Dict]:
        """Get story by ID from stories section"""
        for story in self.data.get('stories', []):
            if story.get('id') == story_id:
                return story
        return None

    def update_story_status(self, story_id: str, new_status: str) -> bool:
        """Update story status"""
        updated = False
        for story in self.data.get('stories', []):
            if story.get('id') == story_id:
                old_status = story.get('status')
                story['status'] = new_status
                print(f"✅ Updated {story_id}: {old_status} → {new_status}")
                updated = True
                break
        
        if updated:
            self._save()
        
        return updated

    def add_story(self, story_data: Dict) -> bool:
        """Add a new story to the stories section"""
        story_id = story_data.get('id')
        if not story_id:
            print("❌ Story must have an 'id' field")
            return False
        
        # Check if story already exists
        if self.get_story(story_id):
            print(f"❌ Story {story_id} already exists")
            return False
        
        self.data['stories'].append(story_data)
        
        # Add to epic's story list if epic is specified
        epic_id = story_data.get('epic')
        if epic_id:
            epic = self.get_epic(epic_id)
            if epic:
                if story_id not in epic.get('stories', []):
                    epic['stories'].append(story_id)
        
        self._save()
        print(f"✅ Added story {story_id}")
        return True

    def list_stories_by_status(self, status: str) -> List[Dict]:
        """List all stories with a specific status"""
        stories = []
        for story in self.data.get('stories', []):
            if story.get('status') == status:
                stories.append({
                    'id': story.get('id'),
                    'title': story.get('title'),
                    'status': story.get('status'),
                    'epic': story.get('epic')
                })
        return stories

    def list_stories_by_epic(self, epic_id: str) -> List[Dict]:
        """List all stories belonging to an epic"""
        stories = []
        for story in self.data.get('stories', []):
            if story.get('epic') == epic_id:
                stories.append({
                    'id': story.get('id'),
                    'title': story.get('title'),
                    'status': story.get('status'),
                    'epic': story.get('epic')
                })
        return stories

    def update_statistics(self):
        """Recalculate statistics from actual data"""
        total_stories = len(self.data.get('stories', []))
        status_counts = {'done': 0, 'in-progress': 0, 'planning': 0, 'blocked': 0}
        
        for story in self.data.get('stories', []):
            status = story.get('status', '').lower()
            if status in status_counts:
                status_counts[status] += 1
        
        epic_status_counts = {'done': 0, 'in-progress': 0, 'planning': 0, 'paused': 0}
        for epic in self.data.get('epics', []):
            status = epic.get('status', '').lower()
            if status in epic_status_counts:
                epic_status_counts[status] += 1
        
        print(f"✅ Statistics: {len(self.data.get('epics', []))} epics, {total_stories} stories")
        print(f"   Epics: {epic_status_counts}")
        print(f"   Stories: {status_counts}")

    def show_summary(self):
        """Display a summary of the current sprint status"""
        epics = self.data.get('epics', [])
        stories = self.data.get('stories', [])
        
        print("\n📊 Sprint Status Summary")
        print("=" * 60)
        print(f"Total Epics: {len(epics)}")
        print(f"Total Stories: {len(stories)}")
        print()
        
        # Stories by status
        status_counts = {}
        for story in stories:
            status = story.get('status', 'unknown')
            status_counts[status] = status_counts.get(status, 0) + 1
        
        print("Stories by Status:")
        for status, count in sorted(status_counts.items()):
            print(f"  {status.title()}: {count}")
        print()
        
        # Epics by status
        epic_status_counts = {}
        for epic in epics:
            status = epic.get('status', 'unknown')
            epic_status_counts[status] = epic_status_counts.get(status, 0) + 1
        
        print("Epics by Status:")
        for status, count in sorted(epic_status_counts.items()):
            print(f"  {status.title()}: {count}")
        print()


def main():
    parser = argparse.ArgumentParser(description='Sprint Status YAML Manager (Simplified)')
    subparsers = parser.add_subparsers(dest='command', help='Commands')
    
    # Summary command
    subparsers.add_parser('summary', help='Show sprint status summary')
    
    # Update story status
    update_parser = subparsers.add_parser('update-story', help='Update story status')
    update_parser.add_argument('story_id', help='Story ID (e.g., story-01)')
    update_parser.add_argument('status', choices=['planning', 'in-progress', 'done', 'blocked'], help='New status')
    
    # Add story
    add_parser = subparsers.add_parser('add-story', help='Add a new story')
    add_parser.add_argument('story_id', help='Story ID (e.g., story-21)')
    add_parser.add_argument('title', help='Story title')
    add_parser.add_argument('epic', help='Epic ID (e.g., epic-auth)')
    add_parser.add_argument('--status', default='planning', choices=['planning', 'in-progress', 'done', 'blocked'], help='Initial status')
    
    # List stories
    list_parser = subparsers.add_parser('list-stories', help='List stories')
    list_parser.add_argument('--status', help='Filter by status')
    list_parser.add_argument('--epic', help='Filter by epic ID')
    
    # Update statistics
    subparsers.add_parser('update-stats', help='Recalculate and show statistics')
    
    args = parser.parse_args()
    
    if not args.command:
        parser.print_help()
        return
    
    try:
        manager = SprintStatusManager()
        
        if args.command == 'summary':
            manager.show_summary()
        
        elif args.command == 'update-story':
            if manager.update_story_status(args.story_id, args.status):
                manager.update_statistics()
            else:
                print(f"❌ Story not found: {args.story_id}")
                sys.exit(1)
        
        elif args.command == 'add-story':
            story_data = {
                'id': args.story_id,
                'title': args.title,
                'status': args.status,
                'epic': args.epic
            }
            if manager.add_story(story_data):
                manager.update_statistics()
            else:
                sys.exit(1)
        
        elif args.command == 'list-stories':
            if args.status:
                stories = manager.list_stories_by_status(args.status)
            elif args.epic:
                stories = manager.list_stories_by_epic(args.epic)
            else:
                print("Please specify --status or --epic")
                sys.exit(1)
            
            print(f"\n📋 Stories ({len(stories)} found):")
            print("=" * 80)
            for story in stories:
                print(f"{story['id']}: {story['title']} [{story['status']}] (Epic: {story['epic']})")
        
        elif args.command == 'update-stats':
            manager.update_statistics()
    
    except Exception as e:
        print(f"❌ Error: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)


if __name__ == '__main__':
    main()
