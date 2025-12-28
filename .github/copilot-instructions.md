# GitHub Copilot Instructions for BMad Method

This workspace uses the BMad Method for software development with multiple specialized AI agents.

## Custom Agents Available

The following custom agents are defined in `.github/chatmodes/`:

- **analyst** - Business Analyst for requirements and PRDs
- **pm** - Product Manager for product planning
- **architect** - Software Architect for system design
- **dev** - Developer for implementation
- **quick-flow-solo-dev** - Quick development workflow
- **sm** - Scrum Master for agile processes
- **tea** - Test Engineer & Automation
- **tech-writer** - Technical documentation
- **ux-designer** - UX/UI design
- **bmad-master** - Master orchestrator

## Agent Usage

Each agent is specialized for specific tasks and follows BMad Method workflows defined in `.bmad/bmm/workflows/`.

## How to Use Agents in GitHub Copilot Chat

1. Open GitHub Copilot Chat panel (Ctrl+Alt+I or Cmd+Alt+I)
2. Click the mode selector (top of chat panel)
3. Select one of the BMad agents from the dropdown
4. The agent will be active for that chat session

## Alternative Usage

You can also mention agents directly in your messages:
- "Using the analyst agent, help me create a PRD"
- "@workspace Use the architect agent for system design"

## Documentation Location

- Full agent definitions: `.bmad/bmm/agents/`
- Workflow configurations: `.bmad/bmm/workflows/`
- Project documentation: `docs/`
- Sprint artifacts: `docs/sprint-artifacts/`
- Coding standards: `docs/standards/coding-standards.md`
- Testing standards: `docs/standards/testing-standards.md`
- API design standards: `docs/standards/api-design.md`
