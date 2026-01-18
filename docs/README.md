# Documentation Structure
# apprun BaaS Platform

This directory contains all project documentation following the **BMad Method** workflow.

---

## 📋 Documentation Hierarchy

```
PRD (Product Requirements)
  ↓
Epics (Business Features)
  ↓
Stories (Implementation Tasks)
  ↓
Sprints (Time-boxed Delivery)
  ↓
Standards (Technical Guidelines)
```

---

## 📁 Directory Structure

| Directory | Purpose | Scope | Owner |
|-----------|---------|-------|-------|
| **[analysis/](./analysis/)** | Product discovery & research | Business requirements gathering | Product Manager |
| **[architecture/](./architecture/)** | System design & tech decisions | Architecture diagrams, ADRs | Architect |
| **[epics/](./epics/)** | **Epic definitions (Single Source of Truth)** | Business-level feature breakdown | Product Manager + Architect |
| **[standards/](./standards/)** | Technical specifications | Coding rules, API design, testing | Architect + Dev Lead |
| **[sprint-artifacts/](./sprint-artifacts/)** | Sprint planning & tracking | Stories, tasks, retrospectives | Scrum Master + Team |
| **[poc/](./poc/)** | Proof of concepts | Validation & experiments | Tech Lead |

---

## 🔄 Workflow Relationship

### **1. PRD → Epics → Stories**
- **[PRD](./prd.md)** defines "what to build" (product vision and functional requirements)
- **[Epics](./epics/)** break down PRD into user-value-focused business features
  - 📌 **Single Source of Truth**: Each epic has its own file in `epics/` folder
  - Example: `epics/1-infrastructure-epic.md`, `epics/5-auth-epic.md`
- **[Stories](./sprint-artifacts/)** decompose Epics into implementable tasks
  - Located in `sprint-artifacts/sprint-N/` folders
  - Example: `sprint-artifacts/sprint-0/1-1-docker-environment.md`

### **2. Stories → Sprints**
- **Sprints** group Stories into 2-week iterations
- Each Sprint delivers working software

### **3. Standards → Implementation**
- **Standards** define "how to build" (technical guidelines)
- Standards are **implemented during Sprints** through Stories
- Example: `standards/api-design.md` → Sprint-0 Story 1 (Response Package)

---

## 📖 Key Documents

- **[prd.md](./prd.md)** - Product Requirements Document
- **[standards/README.md](./standards/README.md)** - Technical standards index
- **[sprint-artifacts/README.md](./sprint-artifacts/README.md)** - Sprint tracking
- **[architecture/tech-architecture.md](./architecture/tech-architecture.md)** - System architecture

---

**Last Updated**: 2025-12-26  
**Maintained By**: Architect Agent
