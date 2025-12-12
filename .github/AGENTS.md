# BMad Method Agents for GitHub Copilot

This workspace includes custom BMAD agents located in `.github/chatmodes/`.

## Available Agents

To use these agents in GitHub Copilot Chat, try:

1. Open GitHub Copilot Chat panel (Ctrl+Alt+I or Cmd+Alt+I)
2. Click the mode selector (top of chat panel)
3. Select one of the BMad agents from the dropdown
4. The agent will be active for that chat session

## Available Chat Modes

- **analyst** - Business Analyst for requirements and PRDs
- **architect** - Software Architect for system design
- **bmad-master** - Master orchestrator for BMad Method workflows
- **dev** - Developer for implementation
- **pm** - Product Manager for product planning
- **quick-flow-solo-dev** - Quick development workflow agent
- **sm** - Scrum Master for agile processes
- **tea** - Test Engineer & Automation
- **tech-writer** - Technical documentation
- **ux-designer** - UX/UI designer

## Alternative Usage

You can also mention agents directly in your messages:
- "Using the analyst agent, help me create a PRD"
- "@workspace Use the architect agent for system design"

## Note

If agents don't appear in the mode selector:
1. Reload VS Code window (Ctrl+Shift+P > "Developer: Reload Window")
2. Check GitHub Copilot Chat extension is enabled and updated
3. Verify you're using the latest version of GitHub Copilot extensions
4. GitHub Copilot chat modes feature may require GitHub Copilot for Business or Enterprise

## Alternative: Use Claude Code

BMad agents are also available in Claude Code as slash commands:
- `/bmad:bmm:agents:analyst`
- `/bmad:bmm:agents:pm`
- `/bmad:bmm:agents:architect`
- etc.
