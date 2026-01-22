# Micro-Stakes API Constitution

## Core Principles

### I. Feature-Based Architecture (NON-NEGOTIABLE)

Every feature MUST be organized as a self-contained module with clear business boundaries within `internal/`. Each feature MUST include exactly these components:
- `model.go` - Domain entities and business models
- `service.go` - Business logic layer
- `repository.go` - Repository interface (abstraction)
- `postgres_repo.go` - PostgreSQL implementation of repository
- `handler.go` - HTTP handlers (Gin controllers)
- `data.go` - Request/Response DTOs
- `errors.go` - Feature-specific error types

**Rationale**: Feature isolation enables parallel development, independent testing, clear ownership, and simplified reasoning about domain boundaries. The prescribed structure enforces separation of concerns while maintaining consistency across the codebase.

### II. Repository Pattern (REQUIRED)

All data access MUST go through repository interfaces defined in `repository.go`. Services MUST depend on repository interfaces, NOT concrete implementations. PostgreSQL implementations live in `postgres_repo.go`.

**Rationale**: Repository abstraction decouples business logic from data access details, enables easier testing with mocks, and allows database implementation changes without touching business logic.

### III. Clean Dependency Flow

Dependencies MUST flow inward: `handler.go` → `service.go` → `repository.go` (interface) ← `postgres_repo.go` (implementation). Handlers depend on services, services depend on repository interfaces, implementations satisfy interfaces.

**Rationale**: This dependency structure ensures business logic remains independent of HTTP transport and database implementation, making the codebase testable, maintainable, and adaptable to technology changes.

### IV. DTO Separation

External API contracts (requests/responses) MUST be defined as DTOs in `data.go`, separate from domain models in `model.go`. Handlers convert between DTOs and domain models. Never expose internal models directly via API.

**Rationale**: Separating API contracts from domain models provides API versioning flexibility, prevents internal refactoring from breaking clients, and allows independent evolution of domain logic and API surface.

### V. Keycloak Authentication Integration

All protected endpoints MUST authenticate via Keycloak using Direct Access Grant flow. Authentication logic MUST be centralized in middleware. Services receive validated user context from handlers.

**Rationale**: Centralized authentication via industry-standard OAuth2/OIDC provider ensures security, auditability, and reduces authentication logic duplication across features.

### VI. Error Handling Standards

Each feature MUST define its domain-specific errors in `errors.go`. Errors MUST include context (operation, entity, reason). HTTP handlers MUST map domain errors to appropriate HTTP status codes. Never leak internal errors to clients.

**API Error Response Format**: All HTTP error responses MUST use a consistent JSON format: `{\"error\": \"message\", \"code\": \"ERROR_CODE\"}`. Validation errors MUST return HTTP 400 with detailed field-level error messages. All errors MUST be defined in `errors.go` and mapped to appropriate HTTP status codes in handlers.

**Rationale**: Structured error handling improves debugging, provides better client error messages, and prevents accidental exposure of sensitive internal details. Consistent error response format enables predictable client integrations and easier debugging.

### VII. Testing Requirements

Each feature MUST include:
- Unit tests for service logic (mocking repositories)
- Repository contract tests validating PostgreSQL implementation
- Integration tests for complete HTTP request flows (handler → service → repository)

**Rationale**: Comprehensive testing at each layer ensures correctness, catches regressions early, validates architectural boundaries, and documents expected behavior.

### VIII. Observability & Logging

All critical operations (authentication, bankroll changes, bet registration) MUST be logged in a structured format including timestamp, user ID, operation type, and outcome. Log levels MUST be used appropriately (INFO, WARN, ERROR). Performance metrics (response times, database query times) for key endpoints MUST be tracked. Authentication failures MUST log IP address and timestamp for security monitoring.

**Rationale**: Multi-user financial system requires comprehensive audit trails and monitoring for security and operational insight.

## Technical Standards

### Stack & Tooling

**Language**: Go 1.23+
**Web Framework**: Gin Gonic (routing, middleware, HTTP handling)
**ORM**: GORM (database access, migrations)
**Database**: PostgreSQL (primary data store)
**Authentication**: Keycloak (OAuth2/OIDC provider via Direct Access Grant)

### Project Structure

```
micro-stakes-api/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/                     # Private application code
│   ├── [feature]/               # Feature modules (users, bankrolls, strategies, bets, dashboard)
│   │   ├── model.go             # Domain entities
│   │   ├── service.go           # Business logic
│   │   ├── repository.go        # Repository interface
│   │   ├── postgres_repo.go     # PostgreSQL implementation
│   │   ├── handler.go           # HTTP handlers
│   │   ├── data.go              # DTOs (requests/responses)
│   │   └── errors.go            # Feature-specific errors
│   ├── domain/                   # Value objects and domain-wide types
│   ├── di/                       # Dependency injection setup
│   ├── infrastructure/           # Infrastructure code
│   │   ├── database/            # Database connection, migrations
│   │   ├── auth/                # Keycloak integration
│   │   └── queue/               # Message queue integration (if applicable)
│   └── shared/                   # Shared utilities and components
│       ├── middleware/          # Authentication, logging, error handling
│       ├── config/              # Configuration management
│       ├── logger/              # Logging utilities
│       └── validator/           # Input validation utilities
├── tests/
│   ├── unit/                    # Unit tests (per feature)
│   ├── integration/             # Integration tests (end-to-end)
│   └── contract/                # Repository contract tests
└── specs/                       # Feature specifications
```

### Performance & Scale

- API response time: <200ms p95 for read operations, <500ms p95 for write operations
- Support 1000+ concurrent users
- Database queries MUST use proper indexing (verified via EXPLAIN ANALYZE)
- Pagination REQUIRED for list endpoints (max 100 items per page)

### Security Standards

- All endpoints except health check MUST require authentication
- User MUST only access their own resources (authorization checks in service layer)
- Passwords/tokens NEVER logged
- Input validation REQUIRED on all handler inputs (use Gin binding validation)
- SQL injection prevention via GORM parameterized queries (never raw SQL with string concatenation)

### Database Migration Standards

All database migrations MUST be versioned and reversible. Migration scripts MUST be tested in a development environment before deployment.

**Rationale**: Reduces risk of data loss, enables safe rollbacks, and improves governance and change control.

## Development Workflow

### Feature Development Lifecycle

1. **Specification**: Create feature spec in `specs/[###-feature-name]/spec.md` with user stories, requirements, success criteria
2. **Planning**: Generate implementation plan with research, data model, contracts, tasks
3. **Test-First**: Write failing tests before implementation (unit → integration)
4. **Implementation**: Follow task order, implement per feature structure pattern
5. **Validation**: All tests pass, manual testing via Postman/curl, quickstart validation
6. **Review**: Code review checks architecture compliance, test coverage, error handling

### Branch Strategy

- Feature branches: `###-feature-name` (issue number prefix)
- No direct commits to `main`
- Pull requests REQUIRED with passing tests and architecture validation

### Code Review Checklist

Reviewers MUST verify:
- [ ] Feature follows prescribed structure (7 required files present)
- [ ] Dependencies flow correctly (handler → service → repository interface)
- [ ] DTOs used for API contracts (no direct model exposure)
- [ ] Authentication middleware applied to protected routes
- [ ] Feature-specific errors defined and properly handled
- [ ] Tests present and passing (unit, contract, integration)
- [ ] Repository interface defined, PostgreSQL implementation satisfies it
- [ ] No code duplication across features (extract to shared package if needed)

## Governance

### Amendment Procedure

Constitution changes require:
1. Proposal with rationale documented
2. Impact analysis on existing features
3. Template updates reflecting new principles
4. Version bump following semantic versioning
5. Approval from project maintainer
6. Migration plan for non-compliant code

### Versioning Policy

- **MAJOR** (X.0.0): Backward-incompatible principle changes requiring code refactoring
- **MINOR** (x.Y.0): New principles added, expanded guidance, new mandatory sections
- **PATCH** (x.y.Z): Clarifications, typo fixes, non-semantic refinements

### Compliance Review

All pull requests MUST pass constitution compliance check:
- Architecture patterns followed (feature structure, dependency flow)
- Technical standards met (stack, testing, security)
- Development workflow observed (spec → plan → test → implement)

### Runtime Guidance Documentation

Implementation-specific runtime guidance MUST be documented in the `docs/` directory as features are developed.

**Rationale**: Centralizes operational knowledge and facilitates troubleshooting and support.

Constitution supersedes all other practices. When in conflict, constitution takes precedence.

**Version**: 1.0.0 | **Ratified**: 2026-01-15 | **Last Amended**: 2026-01-15
