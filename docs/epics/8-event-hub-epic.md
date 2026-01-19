# Epic 8: Event Hub
# apprun BaaS Platform

**Epic ID**: epic-8  
**Related PRD**: [FR-EVENT-001](../prd.md#28-event-center)  
**Owner**: Dev Team  
**Status**: Planning  
**Priority**: P1 (Critical)  
**Estimated Effort**: 3-4 weeks

---

## 1. Epic Overview

### 1.1 Business Goal

Provide event-driven communication between microservices through a pub/sub event bus, enabling loose coupling and asynchronous processing.

### 1.2 Core Value

- Decouple service dependencies through event-driven architecture
- Enable asynchronous processing and notification patterns
- Support event persistence and replay for reliability
- Low-latency event delivery for real-time operations

### 1.3 Acceptance Criteria

- [ ] Event publish latency < 10ms
- [ ] Event delivery latency < 100ms
- [ ] Topics can be created and managed
- [ ] Events are persisted and can be replayed (when Redis backend enabled)
- [ ] Multiple subscribers per topic supported
- [ ] Historical event query capability
- [ ] Prometheus metrics exposed for monitoring
- [ ] Event schema validation enforced

---

## 2. Technical Specifications

### 2.1 Architecture Design

**Layered Anti-Corruption Design**:

```
Business Services → Domain Event Layer (Anti-Corruption) → Watermill Adapter → Watermill → Backend
```

**Key Principles**:
- **Domain Event Layer**: Business-facing interfaces (`EventPublisher`, `DomainEvent`)
- **Watermill Adapter**: Technical implementation hiding Watermill details
- **Watermill**: Infrastructure abstraction (already anti-corruption for backends)
- **Backend**: Pluggable (Redis Streams, SQL, Go Channels, Kafka)

**Technology Stack**:
- Framework: Watermill (vendor-agnostic event streaming)
- Default Backend: Go Channels (in-memory, zero external dependencies)
- Optional Backend: Redis Streams (for persistence and multi-instance scenarios)
- Future: SQL or Kafka support via Watermill adapter swap

**Core Components**:
- `EventPublisher` interface: Business event publishing
- `EventSubscriber` interface: Event consumption with typed handlers
- Watermill Adapter: Bridges domain events to Watermill messages
- Event Store: Persistence layer for replay capability

### 2.2 API Design

**External API**:
- Publish: `POST /api/v1/events/{topic}`
- Query History: `GET /api/v1/events/{topic}/history`

**Internal SDK**:
- `events.Publish(ctx, DomainEvent)` - Type-safe event publishing
- `events.Subscribe(eventType, handler)` - Typed subscription

**Event Naming Convention**:
- Format: `{domain}.{entity}.{action}` (e.g., `user.account.created`)
- Lowercase, dot-separated
- Past tense for actions

---

## 3. User Stories

### Story 8.1: Event Publishing

As a **microservice developer**,  
I want **to publish events to a topic**,  
So that **other services can react to state changes asynchronously**.

**Acceptance Criteria:**

**Given** a valid topic name and event payload  
**When** I publish an event via API  
**Then** the event is persisted and delivered to all subscribers  
**And** the publish operation completes in < 10ms

---

### Story 8.2: Event Subscription

As a **microservice developer**,  
I want **to subscribe to topics and receive events**,  
So that **my service can react to events from other services**.

**Acceptance Criteria:**

**Given** a topic subscription is registered  
**When** an event is published to that topic  
**Then** my handler receives the event within 100ms  
**And** the event is delivered at-least-once

---

### Story 8.3: Topic Management

As a **system administrator**,  
I want **to create and manage event topics**,  
So that **I can organize events by domain**.

**Acceptance Criteria:**

**Given** administrative access  
**When** I create/update/delete a topic  
**Then** the operation succeeds with proper validation  
**And** existing subscriptions are notified of changes

---

### Story 8.4: Event History & Replay

As a **system operator**,  
I want **to query historical events and replay them**,  
So that **I can debug issues or recover from failures**.

**Acceptance Criteria:**

**Given** a topic with historical events  
**When** I query event history with time range  
**Then** matching events are returned  
**And** I can trigger replay to re-deliver events

---

### Story 8.5: Observability & Health Monitoring

As a **system operator**,  
I want **to monitor event hub health and performance metrics**,  
So that **I can detect and resolve issues proactively**.

**Acceptance Criteria:**

**Given** the event hub is operational  
**When** I access monitoring endpoints  
**Then** Prometheus metrics are exposed (publish/subscribe rates, latencies, errors)  
**And** failed delivery count is tracked per topic  
**And** dead letter queue captures undeliverable events  
**And** health check endpoint returns system status

---

### Story 8.6: Event Schema Validation

As a **microservice developer**,  
I want **event schemas to be validated at publish time**,  
So that **invalid events are caught early and compatibility is maintained**.

**Acceptance Criteria:**

**Given** a domain event type is defined  
**When** I publish an event  
**Then** the event payload is validated against its schema  
**And** invalid events are rejected with clear error messages  
**And** schema version is tracked in event metadata

---

## 4. Dependencies

- Config system (Epic 3) - for event hub configuration
- Storage service (Epic 6) - potential backend for persistence
- Auth service (Epic 5) - for topic access control

---

## 5. Non-Functional Requirements

- **Performance**: Handle 1000+ events/sec
- **Reliability**: At-least-once delivery guarantee
- **Scalability**: Horizontal scaling via consumer groups
- **Monitoring**: Metrics for publish/subscribe rates and latencies

---

## 6. Decisions & Open Questions

**Decisions Made**:
- ✅ Backend: Go Channels (default), Redis Streams (optional)
- ✅ Schema validation: At publish time using Go struct validation
- ✅ Naming convention: `{domain}.{entity}.{action}` format

**Open Questions**:
- [ ] Multi-tenancy isolation: Namespace per tenant or shared topics with filtering?
- [ ] Event retention policy: How long to keep events when Redis backend is used?
- [ ] Schema evolution: Versioning strategy for breaking changes?

---

## 7. References

- [PRD - Event Center](../prd.md#28-event-center)
- [Watermill Documentation](https://watermill.io/)
- Architecture documents in `docs/architecture/`
