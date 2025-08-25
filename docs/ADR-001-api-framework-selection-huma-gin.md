---
status: "proposed"
date: 2025-01-25
decision-makers: [Joao]
consulted: [Gus, Luis, Alan]
informed: []
---

# API Framework Selection: Huma + Gin for HTTP API Development

## Context and Problem Statement

We are building a new Go-based backend system following hexagonal architecture principles with well-defined use cases containing input validation logic. After developing a proof-of-concept (POC) with gRPC + gRPC Gateway to evaluate protobuf-first development, we need to select the production technology stack for our HTTP API.

During POC evaluation, the development team provided feedback that the gRPC + protobuf approach felt "overengineered" for our REST API requirements. The team specifically stated that "protobufs are gRPC technology and have nothing to do with REST" and expressed concerns about the complexity overhead and cultural fit for straightforward HTTP API development.

However, the POC demonstrated valuable benefits we want to preserve:
- Contract-first API development with compile-time validation
- Auto-generated, synchronized API documentation
- Strong type safety throughout request/response lifecycle
- Clean separation between HTTP concerns and business logic

**Key Challenge**: Select a Go HTTP framework that provides contract-first development benefits while aligning with team cultural preferences and integrating cleanly with our existing hexagonal architecture and use case validation patterns.

## Decision Drivers

* **Team Cultural Alignment** - Framework must feel like natural Go development rather than requiring foreign tooling
* **Contract-First Development** - Maintain schema-driven development with compile-time validation
* **Documentation Generation** - Automatic, synchronized API documentation without manual maintenance
* **Architecture Integration** - Must work seamlessly with existing hexagonal architecture and use case
* **Validation Consolidation** - Avoid duplicating validation logic between HTTP layer and domain layer
* **Development Velocity** - Enable rapid iteration with immediate testing and documentation capabilities
* **Learning Curve** - Minimize onboarding time for team members familiar with Go HTTP patterns
* **Production Readiness** - Framework must be stable and suitable for production deployment
* **CI/CD Integration** - Support automatic documentation generation and endpoint testing workflows

## Considered Options

* **Huma + Gin** - Schema-first framework with automatic OpenAPI generation integrated with mature HTTP router
* **Pure Gin + Swaggo** - Traditional approach with comment-based documentation generation
* **Pure Gin + HTTP Files** - Lightweight approach with manual testing files for MVP development
* **gRPC + gRPC Gateway** - Protobuf-first with dual protocol support (evaluated in POC)
* **Proto definitions + openapi-gen + ogen** - Protobuf-first with manual OpenAPI generation (evaluated in POC) 
* **Gin + Manual OpenAPI** - Hand-maintained OpenAPI specifications with traditional HTTP handlers
* **Markdown documents** - Manual documentation with traditional HTTP handlers 
* **Fiber + Swagger** - Alternative HTTP framework with built-in documentation features

## Decision Outcome

Chosen option: **"Huma + Gin"**, because it uniquely provides contract-first development benefits using Go-native patterns while integrating perfectly with our hexagonal architecture and team preferences.

Huma addresses all our requirements by:
- Using pure Go structs with validation tags instead of protobuf schemas
- Generating comprehensive OpenAPI 3.1 documentation automatically from type definitions
- Providing compile-time validation through Go's type system and interfaces
- Enabling thin HTTP handlers that delegate to existing use cases
- Eliminating documentation drift through code-driven schema generation
- Supporting immediate endpoint testing via built-in Swagger UI

### Consequences

* **Good**, because Go struct-based API definitions align with team's cultural preferences
* **Good**, because automatic OpenAPI generation provides comprehensive documentation without manual maintenance
* **Good**, because thin HTTP layer preserves hexagonal architecture boundaries with existing use cases
* **Good**, because compile-time type safety prevents API contract violations at build time

[//]: # (* **Good**, because validation logic can remain consolidated in domain use case commands)
* **Good**, because Gin integration provides mature, battle-tested HTTP routing and middleware
* **Good**, because built-in Swagger UI enables immediate endpoint testing and API exploration
* **Good**, because single HTTP server simplifies deployment architecture
* **Bad**, because newer framework (Huma) has smaller community compared to pure Gin ecosystem
* **Bad**, because some advanced OpenAPI features may have limitations compared to hand-crafted specifications
* **Neutral**, because requires learning Huma patterns but builds on familiar Go struct and HTTP concepts

### Confirmation

Implementation success will be validated through:
- **Build-time Validation**: Go compiler enforces API contract compliance automatically
- **Documentation Quality**: Generated OpenAPI spec completeness and interactive Swagger UI functionality
- **Architecture Preservation**: Verification that HTTP handlers remain thin with business logic in use cases
- **Team Productivity**: Developer feedback on ease of development compared to alternative approaches
- **CI/CD Integration**: Successful automatic documentation generation and testing in deployment pipeline

## Pros and Cons of the Options

### Huma + Gin (Chosen Option)

Modern schema-first framework with mature HTTP router integration.

* **Good**, because automatic OpenAPI 3.1 generation from Go structs eliminates documentation drift
* **Good**, because compile-time validation through Go type system catches contract violations early
* **Good**, because integrates with mature Gin ecosystem for middleware and routing
* **Good**, because Go struct definitions feel natural and familiar to Go developers
* **Good**, because built-in request/response validation with comprehensive error handling
* **Good**, because supports complex validation scenarios through struct tags and custom validators
* **Good**, because thin HTTP handlers can delegate validation to existing use case commands
* **Good**, because interactive Swagger UI included for immediate endpoint testing
* **Neutral**, because newer framework with growing community but less extensive than pure Gin
* **Bad**, because some OpenAPI advanced features may have implementation limitations
* **Bad**, because adds dependency on Huma framework in addition to Gin

### Pure Gin + Swaggo

Traditional comment-based documentation generation approach.

* **Good**, because built on battle-tested Gin framework with extensive community support
* **Good**, because familiar annotation patterns similar to other language ecosystems
* **Good**, because generates comprehensive OpenAPI documentation with full feature support
* **Good**, because extensive customization options for documentation formatting
* **Neutral**, because established approach with proven track record in Go community
* **Bad**, because comments can drift from implementation creating documentation bugs
* **Bad**, because validation must be duplicated between comments and actual handler code
* **Bad**, because fragile system where code refactoring can break documentation without compile-time detection
* **Bad**, because requires developer discipline to maintain comment accuracy across team

### Pure Gin + HTTP Files (MVP Approach)

Lightweight development-focused approach with minimal tooling overhead.

* **Good**, because minimal learning curve and setup complexity
* **Good**, because `.http` files provide immediate endpoint testing without additional infrastructure
* **Good**, because allows rapid prototyping and MVP development within tight timelines
* **Good**, because complete flexibility in implementation without framework constraints
* **Neutral**, because HTTP files serve as informal documentation adequate for small development teams
* **Bad**, because no structured API documentation for external consumers or team scaling
* **Bad**, because manual validation implementation required for each endpoint
* **Bad**, because significant migration effort required when moving beyond MVP to production system
* **Bad**, because no compile-time contract validation or type safety guarantees

### gRPC + gRPC Gateway (POC Evaluated)

Protobuf-first approach with dual protocol support.

* **Good**, because excellent type safety and schema evolution capabilities
* **Good**, because generates both gRPC and HTTP endpoints from single schema definition
* **Good**, because mature ecosystem with extensive tooling and community support
* **Good**, because superior performance for service-to-service communication scenarios
* **Good**, because comprehensive backward/forward compatibility features
* **Neutral**, because provides more protocol options than currently required
* **Bad**, because team cultural resistance to protobuf as "overengineered" for REST APIs
* **Bad**, because complex toolchain setup with buf, protoc, and multiple code generation steps
* **Bad**, because dual server architecture increases deployment and operational complexity
* **Bad**, because learning curve for protobuf syntax and gRPC concepts slows development velocity

### Gin + Manual OpenAPI

Hand-crafted approach with explicit API specifications.

* **Good**, because complete control over API documentation structure and content quality
* **Good**, because mature Gin framework provides proven production reliability
* **Good**, because OpenAPI specifications can be designed and shared before implementation begins
* **Good**, because supports all OpenAPI features without framework limitations
* **Neutral**, because traditional approach familiar to developers with OpenAPI experience
* **Bad**, because extremely high maintenance overhead for keeping specifications synchronized
* **Bad**, because high probability of documentation/implementation drift over development lifecycle
* **Bad**, because significant time investment required for comprehensive API specification creation
* **Bad**, because validation logic must be implemented separately from documentation definitions

### Fiber + Swagger

Alternative HTTP framework with built-in documentation capabilities.

* **Good**, because fast HTTP framework optimized for performance
* **Good**, because built-in Swagger integration reduces setup complexity
* **Good**, because familiar Express.js-like API patterns for developers with JavaScript background
* **Neutral**, because growing Go community adoption with active development
* **Bad**, because less mature ecosystem compared to Gin for middleware and extensions
* **Bad**, because comment-based documentation approach shares fragility issues with Swaggo
* **Bad**, because team would need to learn new framework patterns instead of leveraging Gin familiarity
* **Bad**, because smaller community support for troubleshooting and best practices

### Proto definitions + openapi-gen + ogen

Alternative protobuf approach with HTTP-only code generation.

* **Good**, because maintains contract-first benefits of protobuf schema definitions
* **Good**, because generates type-safe Go interfaces and validation from OpenAPI specs
* **Good**, because produces comprehensive OpenAPI documentation automatically
* **Good**, because avoids dual server complexity by focusing on HTTP-only output
* **Neutral**, because two-stage generation process (proto → OpenAPI → Go code)
* **Bad**, because still requires protobuf maintenance which team culturally rejects
* **Bad**, because complex toolchain with multiple code generation steps
* **Bad**, because doesn't address core team objection to protobuf for REST APIs
* **Bad**, because additional complexity over simpler Go-native approaches

### Markdown documents

Manual documentation approach with traditional HTTP handlers.

* **Good**, because complete flexibility in documentation structure and presentation
* **Good**, because no framework dependencies or code generation complexity
* **Good**, because familiar documentation format for most developers
* **Good**, because can include detailed examples, tutorials, and complex explanations
* **Neutral**, because traditional approach with established tooling and workflows
* **Bad**, because extremely high maintenance overhead keeping docs synchronized with code
* **Bad**, because no automatic validation of documented API contracts against implementation
* **Bad**, because high probability of documentation drift and inconsistencies over time
* **Bad**, because no interactive testing capabilities or automated client generation

## More Information

### Architecture Integration Strategy

Our implementation will maintain hexagonal architecture principles through clear layer responsibilities:

**HTTP Layer (Huma + Gin)**: Handles protocol-specific concerns including request parsing, response serialization, routing, and OpenAPI documentation generation. Performs only structural validation needed for JSON unmarshaling.

**Handler/Translation Layer**: Thin functions that convert HTTP requests into use-case input objects and translate domain responses back to HTTP. No business validation - purely mechanical translation between layers.

**Domain Layer**: Existing use case commands remain the authoritative source for all business validation and rules. Inputs like `SignupUseCaseInput` continue to encapsulate comprehensive validation using struct tags and the go-playground/validator library.

This approach ensures single source of truth for validation while enabling automatic documentation generation from HTTP layer structure.

### Risk Assessment and Mitigation

**Framework Maturity Risk**: Huma is newer than pure Gin
- *Mitigation*: Huma integrates with Gin's mature routing, limiting risk to documentation generation features

**Learning Curve Risk**: Team needs to learn Huma patterns
- *Mitigation*: Huma builds on familiar Go structs and HTTP concepts, minimal new concepts required