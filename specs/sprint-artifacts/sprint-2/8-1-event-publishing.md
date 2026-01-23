# Story 8.1: Event Publishing

Status: ready-for-dev

## Story

As a **microservice developer**,
I want **to publish domain events to topics using a type-safe interface**,
so that **other services can react to state changes asynchronously without tight coupling**.

## Acceptance Criteria

1. **AC1**: Event publishing completes in < 10ms (P95 latency measured via benchmark test)
2. **AC2**: Go Channel backend successfully publishes events (Redis backend deferred to Story 8.4)
3. **AC3**: Type-safe domain event interface enforces: non-empty EventType, non-nil EventData, valid Timestamp
4. **AC4**: Event metadata includes timestamp, event type, and correlation ID
5. **AC5**: Failed publishes return clear error messages
6. **AC6**: Publisher can switch backends (Go Channel ↔ Redis) via configuration (verified via integration test with different configs)
7. **AC7**: Backend unavailability returns context-aware error within 100ms (no hanging)

## Definition of Done

- [ ] All 7 acceptance criteria verified and passing
- [ ] All tasks and subtasks completed and checked off
- [ ] Unit test coverage ≥80% for `modules/events` module
- [ ] Integration tests passing in CI
- [ ] Performance benchmark documented (P50/P95/P99 latencies)
- [ ] Code reviewed and approved by team
- [ ] Documentation updated (usage examples, API docs)
- [ ] No new linting errors introduced
- [ ] Changes merged to main branch

## Tasks / Subtasks

- [ ] Task 1: Domain Event Layer (AC: #3, #4)
  - [ ] Define `DomainEvent` interface with EventType(), EventData(), Timestamp()
  - [ ] Define `EventPublisher` interface with Publish(ctx, event) method
  - [ ] Create sample business events (e.g., UserCreatedEvent, OrderPlacedEvent)
  - [ ] Add correlation ID generation utility

- [ ] Task 2: Watermill Adapter Implementation (AC: #1, #2)
  - [ ] Implement `WatermillPublisher` struct
  - [ ] Bridge DomainEvent → Watermill Message conversion
  - [ ] Implement Go Channel backend publisher factory
  - [ ] Add backend interface for future Redis support (factory pattern)

- [ ] Task 3: Configuration & Initialization (AC: #6)
  - [ ] Add event hub config section in config system
  - [ ] Define backend type field (default: gochan)
  - [ ] Create publisher initialization in server startup
  - [ ] Add graceful shutdown for publisher cleanup

- [ ] Task 4: Testing (AC: #1, #3, #5, #7)
  - [ ] Unit tests for DomainEvent implementations (test validation rules)
  - [ ] Unit tests for WatermillPublisher with mock backend
  - [ ] Integration test: Publish event via Go Channel backend
  - [ ] Integration test: Load different backend configs (AC#6 verification)
  - [ ] Performance benchmark: Verify <10ms P95 latency (1000 events)
  - [ ] Error handling tests: Invalid events (nil/empty fields), backend unavailability

- [ ] Task 5: Documentation (AC: All)
  - [ ] Add usage example in README or docs/
  - [ ] Document event naming convention: `{domain}.{entity}.{action}`
  - [ ] Document how to add new domain events
  - [ ] API documentation for EventPublisher interface

## Dev Notes

### Architecture Constraints

- **Layered Anti-Corruption Design**: Business code depends ONLY on `EventPublisher` interface, not Watermill
- **Backend Pluggability**: Use factory pattern to select backend at runtime
- **Zero External Dependencies (Default)**: Go Channels backend requires no Redis/Kafka
- **Type Safety**: All events are Go structs, not raw `[]byte` or JSON strings

### Technology Stack

- **Framework**: Watermill v1.x (vendor-agnostic event streaming)
- **Default Backend**: Go Channels (in-memory, single-instance)
- **Optional Backend**: Redis Streams (persistence, multi-instance)
- **Libraries**:
  - `github.com/ThreeDotsLabs/watermill`
  - `github.com/ThreeDotsLabs/watermill-redisstream` (optional)
  - `github.com/redis/go-redis/v9` (optional)

### File Structure

```
core/
  modules/
    events/                    # Event Hub module
      publisher.go             # EventPublisher interface, DomainEvent interface
      types.go                 # Sample event types (UserCreatedEvent, etc.)
      watermill_adapter.go     # WatermillPublisher implementation
      factory.go               # Backend factory (NewEventPublisher)
  config/
    default.yaml              # Add event_hub section
```

### Testing Requirements

- **Unit Test Coverage**: >80% for events module (core/modules/events)
- **Integration Tests**: Test Go Channel backend with real Watermill instance
- **Performance Benchmark**: Publish 1000 events, measure P50/P95/P99 latencies
- **Mock Publisher**: Create for testing business logic without real event bus

### References

- Epic: [docs/epics/8-event-hub-epic.md](../epics/8-event-hub-epic.md)
- PRD: [docs/prd.md#28-event-center](../prd.md)
- Watermill Docs: https://watermill.io/
- Event Naming Convention: `{domain}.{entity}.{action}` (lowercase, past tense)

## Dev Agent Record

### Agent Model Used

_To be filled by dev agent_

### Debug Log References

_To be filled during implementation_

### Completion Notes List

_To be filled during implementation_

### File List

_To be filled during implementation_
