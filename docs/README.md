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
| **[standards/](./standards/)** | Technical specifications | Coding rules, API design, testing | Architect + Dev Lead |
| **[sprint-artifacts/](./sprint-artifacts/)** | Sprint planning & tracking | Stories, tasks, retrospectives | Scrum Master + Team |
| **[poc/](./poc/)** | Proof of concepts | Validation & experiments | Tech Lead |

---

## 🔄 Workflow Relationship

### **1. PRD → Epics → Stories**
- **PRD** defines "what to build" (product vision)
- **Epics** break down PRD into business features
- **Stories** decompose Epics into implementable tasks

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
- **[database-migrations.md](./database-migrations.md)** - Database migration guide (Atlas)

---

**Last Updated**: 2025-12-26  
**Maintained By**: Architect Agent
