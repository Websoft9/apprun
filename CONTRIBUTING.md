# Contributing to apprun

Thank you for your interest in contributing to **apprun**! This project follows the **BMad Method** for AI-assisted development.

---

## 🚀 Getting Started

### Prerequisites
- Go 1.21+
- Git
- GitHub account
- **AI Coding Agent** (GitHub Copilot, Cursor, or similar)

---

## 📋 BMad Method Workflow

We use the **BMad Method** for AI-assisted development. Follow these steps:

### 1. **Understand the Documentation**
```bash
# Read project standards
docs/standards/README.md      # Technical standards index
docs/standards/coding-standards.md
docs/standards/api-design.md
```

### 2. **Find a Task**
- Check [Sprint Artifacts](./docs/sprint-artifacts/) for current sprint stories
- Look for issues tagged `good-first-issue` or `help-wanted`
- Review [Sprint-0 Stories](./docs/sprint-artifacts/sprint-0/stories.md)

### 3. **AI Coding with BMad Method**

#### **Step 1: Load Context into AI Agent**
```bash
# Share relevant documentation with your AI agent
@workspace /docs/standards/coding-standards.md
@workspace /docs/standards/api-design.md
@workspace /docs/sprint-artifacts/sprint-0/stories.md
```

#### **Step 2: Use AI Agent for Implementation**
- **Ask specific questions**: "How should I implement response package according to standards?"
- **Request code generation**: "Generate handler following api-design.md Section 4"
- **Verify compliance**: "Does this code follow coding-standards.md Section 10?"

#### **Step 3: Iterate with AI Guidance**
- AI reviews code against project standards
- AI suggests improvements based on BMad documentation
- AI helps write tests following testing-standards.md

### 4. **Development Process**
```bash
# Create feature branch
git checkout -b feature/story-1-response-package

# Implement with AI assistance
# - Follow docs/standards/coding-standards.md
# - Reference story acceptance criteria
# - Write tests (coverage > 80%)

# Run tests
make test

# Run linter
make lint

# Commit with conventional format
git commit -m "feat(response): implement unified response package

- Add Success/Error/List functions
- Follow api-design.md response format
- Unit tests with 90% coverage

Ref: Sprint-0 Story 1"
```

### 5. **Submit Pull Request**
- Title: `[Story-X] Brief description`
- Description: Reference story, include checklist
- Link to relevant documentation
- Request AI-assisted code review

---

## 🤖 AI Coding Best Practices

1. **Always load project standards first** - AI agents perform better with context
2. **Reference specific sections** - Point AI to exact standards (e.g., "Section 10.4")
3. **Validate with AI** - Ask AI to check compliance before committing
4. **Use AI for test generation** - AI can generate comprehensive test cases

---

## 📚 Key Documentation

- [Development Process](./docs/standards/devops-process.md)
- [Code Review Checklist](./docs/standards/devops-process.md#33-code-review-清单)
- [Commit Message Format](./docs/standards/devops-process.md#22-commit-message-规范)

---

## 🙋 Need Help?

- Open a [Discussion](https://github.com/Websoft9/apprun/discussions)
- Join our community chat
- Tag maintainers in issues

---

**Thank you for contributing!** 🎉
