---
stepsCompleted: []
inputDocuments: 
  - "docs/analysis/product-brief.md"
documentCounts:
  briefs: 1
  research: 0
  brainstorming: 0
  projectDocs: 0
workflowType: 'prd'
lastStep: 0
project_name: 'apprun'
user_name: 'Root'
date: '2025-12-24'
status: 'active'
version: '5.0-bmad-compliant'
---

# Product Requirements Document - apprun (BMad v5.0)

**Author:** Root  
**Date:** 2025-12-24  
**Status:** Active  
**Related Documents:** [Product Brief](./analysis/product-brief.md)

---

## 1. Document Description

### 1.1 Document Purpose and Simplification Principles

This PRD defines the **Functional Requirements (FR)** and **Non-Functional Requirements (NFR)** of the apprun BaaS platform.

**BMad Methodology Compliance**:
1. **Separation of Concerns**: PRD defines "what capabilities are needed", not prescribing "what technology to implement with"
2. **Deferred Technical Decisions**: Specific technology selections are postponed to the technical architecture phase
3. **Focus on Acceptance Criteria**: Clear, testable business value and acceptance criteria

**Simplification Principles**:
1. **Capability-Oriented**: Describe needed capabilities, not implementation approaches
2. **Consolidate Detail Requirements**: Related detailed functional requirements are merged into a single FR
3. **Clear Constraints**: Provide implementation constraints to guide architectural decisions
4. **Maintain Flexibility**: Technology selection is decided during the architecture phase based on actual circumstances

**Relationship with Product Brief**:

- **Product Brief**: Defines business vision, market positioning, 13 core modules
- **PRD (This Document)**: Defines verifiable functional requirements, distinguishes self-developed vs integrated

### 1.2 Target Audience

- Architects: Technology selection and integration solution design
- Development Team: Self-developed module implementation
- Testing Team: Integration testing and acceptance criteria
- Product Manager: Progress tracking and scope management

---

## 2. Core Module Functional Requirements

### 2.1 Authentication & Authorization

#### FR-AUTH-001: Authentication & Authorization Service

- **Requirement**: Provide complete user authentication, authorization, and team collaboration capabilities
- **Core Capabilities**:
  - User registration, login, logout (Email + Password)
  - Secure password storage (hash encryption)
  - JWT Token authentication (Access Token + Refresh Token)
  - Password strength validation and modification
  - Project-based team collaboration and permission management
  - RBAC role permission management
- **Acceptance Criteria**:
  - Users can register and login via API (Email + Password)
  - Passwords securely stored (irreversible hashing)
  - JWT Token correctly issued and verified
  - Token verification integrated into API middleware
  - Users can join multiple Projects with proper permission isolation between Projects
  - Permission verification works correctly (unauthorized access returns 403)
  - API authentication response time P95 < 10ms
- **Implementation Constraints**:
  - MVP Phase: Use Go Native Auth (bcrypt + JWT + Casbin)
  - Production Phase: Consider migration to enterprise-grade authentication service (Ory Kratos, etc.)

---

### 2.2 Data Modeling

#### FR-DATA-001: Data Model Definition and Management
- **Requirement**: Support DSL or configuration-based data model definition, auto-generate CRUD API
- **Core Capabilities**:
  - Field type definition (String, Integer, Boolean, JSON, DateTime, UUID, etc.)
  - Field constraints (required, unique, length limit, default value)
  - Relationship definition (one-to-many, many-to-many, self-referential)
  - Index definition (single column, composite index)
  - Schema migration management (versioning, rollback)
- **Acceptance Criteria**:
  - Users can define data models via configuration files
  - Models automatically generate database tables and RESTful API after definition
  - API supports pagination, filtering, sorting, and relational queries
  - API response time P95 < 100ms
  - Schema changes can generate migration scripts and support rollback

---

### 2.3 Configuration Center

#### FR-CONFIG-001: Configuration Management Service
- **Requirement**: Centralized configuration storage, management, and dynamic updates
- **Core Capabilities**:
  - File and Key-Value database configuration storage
  - Configuration priority
  - Dynamic configuration updates (no restart required)
- **Acceptance Criteria**:
  - Configuration can be CRUD via API
  - Third-party microservice configuration changes are published to event center in real-time
  - Configuration change history is queryable

---

### 2.4 Function Service

#### FR-FUNC-001: Function Execution Service
- **Requirement**: Support user-defined function deployment and execution
- **Core Capabilities**:
  - Support golang language
  - HTTP trigger execution
  - Resource limits (CPU, memory, timeout)
  - Function log collection
- **Acceptance Criteria**:
  - Users can upload function code and deploy
  - Functions can be invoked via HTTP
  - Function executions are mutually isolated
  - Function logs are queryable

---

### 2.5 Plugin Extension

#### FR-PLUG-001: Plugin Extension Mechanism
- **Requirement**: Provide system-level plugin extension capability, support non-invasive system customization
- **Core Capabilities**:
  - Based on RPC protocol
  - Plugin lifecycle management (load, execute, unload)
  - Plugin security isolation and permission control
  - Plugin configuration and version management
- **Acceptance Criteria**:
  - Plugins can be loaded and executed through standard interfaces
  - Plugins are securely isolated from host system
  - Plugin configuration can be dynamically updated
  - Plugin version compatibility is guaranteed

---

### 2.6 File Storage Service

#### FR-STORAGE-001: Unified File Storage Service
- **Requirement**: Provide unified file storage and folder management capability, support object storage mounting
- **Core Capabilities**:
  - File upload, download, delete, move
  - Folder create, delete, rename
  - File list query (support folder structure, pagination, filtering)
  - Object storage mounting (flat structure mapped to folder hierarchy)
  - Generate file access URL
  - Chunked upload (large files >10MB)
- **Acceptance Criteria**:
  - Files can be organized by folder (/project1/docs/readme.txt)
  - Folder operations consistent with local file system
  - Object storage transparently mounted (users unaware of underlying implementation)
  - File permissions inherited from folders

---

### 2.7 Workflow Service

#### FR-WORKFLOW-001: Workflow Engine
- **Requirement**: Provide reliable workflow orchestration and execution capability
- **Core Capabilities**:
  - Workflow definition (YAML/code)
  - Multiple trigger methods (event, scheduled, manual)
  - Built-in nodes (HTTP, SMTP, database, etc.)
  - Custom node extension
- **Acceptance Criteria**:
  - Workflows can be defined and triggered
  - Workflow execution status is queryable
  - Automatic retry on failure (configurable)
  - Cron scheduled tasks execute on schedule

---

### 2.8 Event Center

#### FR-EVENT-001: Event Bus Service
- **Requirement**: Provide event publish/subscribe capability between microservices
- **Core Capabilities**:
  - Event publish/subscribe (Pub/Sub)
  - Topic management
  - Event persistence and replay
- **Acceptance Criteria**:
  - Event publish latency < 10ms
  - Event delivery latency < 100ms
  - Historical events can be queried and replayed

---

### 2.9 Real-time Data Push

#### FR-REALTIME-001: Real-time Push Service
- **Requirement**: Provide server-to-client real-time data push capability
- **Core Capabilities**:
  - WebSocket connection management
  - Server-initiated push
  - Data change push (Database CDC)
- **Acceptance Criteria**:
  - Clients can establish WebSocket connections
  - Push latency < 100ms
  - Data changes automatically pushed to subscribed clients

---

### 2.10 Internationalization

#### FR-I18N-001: Internationalization Service
- **Requirement**: Multi-language content management and switching
- **Core Capabilities**:
  - Key-Value translation storage
  - Support multiple languages (Chinese, English, Japanese, etc.)
  - Variable interpolation
  - Language switching based on HTTP Header/Cookie
- **Acceptance Criteria**:
  - Translation content can be CRUD
  - API returns corresponding translations based on language
  - Default language configuration takes effect

---

### 2.11 Logging & Monitoring

#### FR-LOG-001: Logging and Monitoring Service
- **Requirement**: Provide centralized log collection and system monitoring capability
- **Core Capabilities**:
  - Centralized log collection (unified format, Trace ID)
  - Log query (time range, keyword, Trace ID)
  - Basic performance metrics (CPU, memory, API response time, QPS)
  - Alert rules and notifications
- **Acceptance Criteria**:
  - All module logs can be queried uniformly
  - Log query response < 5 seconds
  - Performance metrics viewable in real-time
  - Alert latency < 1 minute

**MVP Does Not Include**:
- ❌ Distributed Tracing
- ❌ Custom Dashboard
- ❌ AI-driven anomaly detection

---

### 2.12 API Gateway

#### FR-GATEWAY-001: API Gateway Service
- **Requirement**: Provide unified API entry and routing capability, and Reverse Proxy
- **Core Capabilities**:
  - Route forwarding (path-based)
  - Authentication integration (JWT verification)
  - Permission checking (RBAC)
  - Request logging (Trace ID)
  - Reverse Proxy
- **Acceptance Criteria**:
  - Based on Reverse Proxy, requests correctly forwarded to backend services
  - Unauthenticated requests return 401
  - Unauthorized requests return 403
  - All requests have log records

**MVP Does Not Include**:
- ❌ Rate Limiting
- ❌ Circuit Breaker
- ❌ Service Discovery

---

### 2.14 License Management

#### FR-LICENSE-001: License Management Service
- **Requirement**: Provide system feature toggle and License verification capability
- **Core Capabilities**:
  - License generation and verification
  - Feature toggle management
  - Usage monitoring
- **Acceptance Criteria**:
  - System loads License configuration on startup
  - Requests that fail verification return 403
  - License changes can be dynamically updated

---

## 3. Non-Functional Requirements (NFR)

### NFR-001: Performance Requirements
- API response time P95 < 100ms (1000 concurrent)
- System throughput > 10,000 QPS
- Real-time push latency < 100ms (P95)

### NFR-002: Availability Requirements
- System availability > 99.9% (monthly)
- API error rate < 0.1%

### NFR-003: Scalability Requirements
- Support horizontal scaling (stateless service design)
- Modules can be independently deployed and upgraded

### NFR-004: Security Requirements
- Sensitive data encryption (password bcrypt, HTTPS/TLS)
- SQL injection protection (parameterized queries)
- All APIs require authentication and authorization

### NFR-005: Maintainability Requirements
- Unit test coverage > 70%
- Complete logging and monitoring system

### NFR-006: Deployability Requirements
- Docker containerized deployment (docker-compose one-click startup, externalized configuration)

### NFR-007: Lightweight Requirements
- Single machine (1C2G) can run core functions, core service memory usage < 512MB, CPU < 50% (idle)

### NFR-008: Development Standards Requirements
- **API Design**: Unified response format, RESTful style, unified error code system, API version management
- **Code Quality**: Naming conventions, layered architecture, dependency injection, unified error handling
- **Testing System**: Unit tests > 70%, integration tests cover core modules, CI/CD integration
- **Reference Documentation**: `docs/standards/` (created during architecture phase)

---


## 4. MVP Scope Description

### 4.1 Features Included in MVP

All FRs and NFRs defined in this PRD are within MVP scope.

### 4.2 Features Not Included in MVP

- ❌ Data backup and recovery (belongs to operations functionality, can be provided through cloud service providers or third-party tools in later versions)
- ❌ Visual workflow orchestrator (Drag & Drop UI)
- ❌ Advanced image editing (filters, watermarks)
- ❌ Video/audio processing
- ❌ Advanced API Gateway features (rate limiting, circuit breaker, service discovery)
- ❌ Distributed Tracing
- ❌ Custom monitoring Dashboard
- ❌ AI/ML integration
- ❌ Native mobile applications
- ❌ Multi-tenant isolation in SaaS mode

### 4.3 Priority Description

Refer to the priority of 13 core modules defined in [Product Brief - MVP Scope](./analysis/product-brief.md#mvp-scope).

---

## 5. Appendix

### 5.1 Glossary

| Term | Full Name | Description |
|------|------|------|
| FR | Functional Requirement | Functional Requirement |
| NFR | Non-Functional Requirement | Non-Functional Requirement |
| RBAC | Role-Based Access Control | Role-Based Access Control |
| JWT | JSON Web Token | JSON Web Token |
| CRUD | Create, Read, Update, Delete | Create, Read, Update, Delete |
| QPS | Queries Per Second | Queries Per Second |
| P95 | 95th Percentile | 95th Percentile |
| CDC | Change Data Capture | Change Data Capture |
| DSL | Domain-Specific Language | Domain-Specific Language |

### 5.2 Reference Documents

- [Product Brief](./analysis/product-brief.md) - Product Vision and Strategy
- [Epics](./epics/) - **Single Source of Truth for Epic Definitions** (Business Feature Breakdown)
- [Technical Architecture](./architecture/) - Technical Architecture Design
- [Waterflow](https://github.com/Websoft9/Waterflow) - Workflow Engine Project