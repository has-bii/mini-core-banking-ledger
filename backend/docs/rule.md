# Backend Architecture & Design Guidelines
**Pattern:** Controller-Service-Repository Architecture
**Scope:** Layer Responsibilities, Data Flow, Error Handling, Validation, and Dependency Injection

---

## 1. Core Architecture Layers

The application is structured around the principle of **Separation of Concerns**, ensuring each component has a distinct, single responsibility.

### 1.1 Controller Layer (The Edge / Presentation)
* **Primary Role:** Entry point for HTTP requests. Handles routing, HTTP protocol details (headers, query strings, body), and response formatting.
* **Responsibilities:**
  * Extract parameters and payloads from incoming requests.
  * Delegate business logic execution to the Service layer.
  * Return appropriate HTTP status codes (`200 OK`, `201 Created`, `400 Bad Request`, `404 Not Found`).
* **Anti-Patterns:** No business logic, no direct database queries, no SQL code.

### 1.2 Service Layer (Business Logic)
* **Primary Role:** The heart of the application; orchestrates core business rules and workflows.
* **Responsibilities:**
  * Execute business logic (calculations, eligibility checks, orchestration).
  * Delegate writes that must succeed together to one atomic repository operation (e.g., creating a user with credentials).
  * Do not receive database clients, transactions, or transaction-starting interfaces such as `database.Beginner`.
  * Convert database entities into clean domain objects or DTOs (stripping sensitive data like password hashes).
* **Anti-Patterns:** No knowledge of HTTP protocols (status codes, headers, req/res objects).

### 1.3 Repository Layer (Data Access)
* **Primary Role:** Acts as the bridge between business logic and data storage (PostgreSQL, MongoDB, etc.).
* **Responsibilities:**
  * Handle CRUD operations using parameterized SQL through pgx; this project deliberately does not use an ORM or query builder.
  * Encapsulate data access logic and persistence details.
  * Own transaction boundaries for atomic persistence operations, including writes across multiple tables. Commit all related writes together or roll them all back.
  * Keep business decisions in services; coordinating related SQL writes is a persistence responsibility.
* **Anti-Patterns:** No business logic or decision-making regarding application rules.

---

## 2. Dependency Injection (DI) & Flow Rules

Dependencies must flow strictly downward. The application's Inversion of Control (IoC) container manages instantiation via constructor injection.

### 2.1 The Allowed Injection Map
* **Controllers** $\rightarrow$ Can inject $\rightarrow$ **Services**
* **Services** $\rightarrow$ Can inject $\rightarrow$ **Repositories** AND other **Services** (with caution)
* **Repositories** $\rightarrow$ Can inject $\rightarrow$ **Database Clients only**

### Transaction Ownership
* Services call a single repository method for writes that must be atomic; they do not start or manage transactions.
* The repository begins, commits, and rolls back the transaction internally using pgx. No separate `Beginner` interface is required by this architecture.
* Related table writes belong in that atomic repository method; repositories do not construct or inject other repositories.
* Example: `AuthService` hashes the password and calls `UserStore.CreateWithCredentials`; the store inserts both the user and credentials in one transaction.
* Never replace an atomic operation with independent writes: a credentials failure must not leave a user partially registered.

### 2.2 Service-to-Service Injection Guidelines
* **When it is allowed:** For shared, cross-cutting domain logic or common workflows (e.g., an `OrderService` injecting a `NotificationService` or `InventoryService`).
* **The Danger Zone (Circular Dependencies):** Avoid letting `Service A` depend on `Service B` while `Service B` depends on `Service A`.
  * *Resolution:* Extract shared logic into a `Service C`, or refactor using an event-driven/event-emitter pattern.

---

## 3. Data Validation Strategy

Validation is distributed across layers to ensure security, syntactic correctness, and data integrity.

| Layer | Validation Type | Description | On Failure |
| :--- | :--- | :--- | :--- |
| **Controller** | Structural / Syntactic | Validates types, required fields, string lengths, and basic formats via DTO validation libraries. | `400 Bad Request` |
| **Service** | Business / State | Validates application logic, entity existence, permissions, and system state constraints. | `409 Conflict`, `422 Unprocessable` |
| **Repository / DB** | Integrity Constraints | Database-level checks such as unique keys, foreign keys, and not-null constraints. | Handled via DB error translation |

---

## 4. Error Handling & Data Flow

### 4.1 Error Lifecycle
1. **Repository:** Catches low-level database failures (connection drops, timeout errors).
2. **Service:** Translates database errors or rule failures into meaningful **Business Exceptions** (e.g., `UserNotFoundException`, `InsufficientBalanceException`).
3. **Controller / Global Handler:** Catches business exceptions and translates them into standard **HTTP status codes and JSON responses**.

### 4.2 Data Flow Lifecycle
* **Upward Flow:** Data flows from database rows $\rightarrow$ Repository (Entities) $\rightarrow$ Service (Domain Models/DTOs) $\rightarrow$ Controller (HTTP JSON Response).
* **Data Sanitization:** The service layer strips out sensitive internal properties (e.g., password hashes, internal audit flags) before passing data upward to the controller.
