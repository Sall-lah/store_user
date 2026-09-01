## ADDED Requirements

### Requirement: Standardized Microservice README Structure
The repository SHALL contain a `README.md` structured according to the standard microservice template, including header badges, table of contents, mermaid architecture overview, key features, technology stack, repository structure, environment variables, database setup, local development guide, API endpoints catalog, Kafka event streaming specifications, Redis rate limiting policies, testing guides, and container deployment instructions.

#### Scenario: Visual inspection of table of contents and structure
- **WHEN** a developer inspects the root `README.md`
- **THEN** all standardized sections (Architecture Overview, Key Features, Tech Stack, Repository Structure, Environment Configuration, Database Setup, Getting Started, API Endpoints, Kafka Domain Events, Rate Limiting, Testing, Docker Deployment) are present with working navigation anchors.

#### Scenario: Architecture overview with Mermaid diagram
- **WHEN** a developer reviews the Architecture Overview section
- **THEN** an up-to-date Mermaid flowchart illustrates API Gateway header offloading (`X-User-Id`, `X-User-Role`, `X-User-Email`), synchronous gRPC pre-flight active order checks to `store_order`, Prisma PostgreSQL persistence (`user_profiles`, `user_notifications`), and Kafka `user.events` publishing to downstream consumers (`store_auth`, `store_order`).

#### Scenario: Comprehensive API and rate limit catalog
- **WHEN** a developer reviews the API Endpoints and Rate Limiting sections
- **THEN** every endpoint (Profile, Account deletion, Admin governance, In-app notifications, Health, Docs) is mapped with its HTTP method, path, authentication headers, purpose, and sliding-window rate limit rules.

#### Scenario: Event streaming specification
- **WHEN** a developer reviews the Kafka Domain Events section
- **THEN** JSON payload schemas and downstream consumption behavior for `user.deleted` and `user.banned` are documented in detail.
