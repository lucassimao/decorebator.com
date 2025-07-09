# Architecture Review Summary

## Architecture Health Score: 6/10

## Key Strengths
- **Clear 3-Tier Architecture**: The project follows a classic 3-tier architecture, which is easy to understand and maintain.
- **Asynchronous Processing**: The use of River for background job processing is a good choice for a scalable and resilient system.
- **Well-Structured Makefile**: The `Makefile` is well-organized and provides a good set of commands for development, testing, and CI/CD.

## Critical Architecture Issues
- **[Security] Weak JWT Signing Algorithm (HS256)**: The use of a symmetric algorithm for JWT signing is a major security risk.
- **[Security] Use of Unmaintained JWT Library**: The project uses a JWT library that is no longer maintained and may contain known vulnerabilities.
- **[Code Smell] Global State**: The use of global variables and `init()` functions for dependency management creates tight coupling and makes the system difficult to test.

## Architecture Improvement Opportunities
- **Adopt a Hexagonal Architecture**: A hexagonal architecture would provide better separation of concerns and make the application more modular and testable.
- **Implement a Robust Dependency Injection Framework**: A dependency injection framework would help to manage dependencies more effectively and reduce coupling.
- **Introduce a Caching Layer**: A caching layer would improve performance and reduce the load on the database.
- **Improve Observability**: The project needs a more comprehensive monitoring and observability strategy, including distributed tracing and more detailed logging.

## Scalability Readiness
- The system is mostly stateless, which is good for horizontal scaling. However, the global River client and the lack of a distributed session store would be problematic in a multi-instance deployment.

## Architecture Metrics
- **Coupling Score**: 7/10 (Lower is better)
- **Cohesion Score**: 5/10 (Higher is better)
- **Testability Score**: 4/10
- **Maintainability Score**: 6/10

# Code Review Summary

## Critical Issues (Fix Immediately)
- **None**

## Major Issues (Address Soon)
- **[Security] Weak JWT Signing Algorithm (HS256)**
  - **Location**: `internal/http/midlewares.go:40`, `internal/service/user.go:96`
  - **Severity**: Major
  - **Description**: The application uses the `jwt.SigningMethodHMAC` (HS256) symmetric algorithm for signing JWTs. This is not recommended for production environments.
  - **Impact**: If the JWT secret key is compromised, an attacker can forge tokens and impersonate any user.
  - **Recommendation**: Switch to an asymmetric algorithm like RS256. This involves generating a public/private key pair and using the private key to sign tokens and the public key to verify them.
- **[Security] Use of Unmaintained JWT Library**
  - **Location**: `go.mod:8`
  - **Severity**: Major
  - **Description**: The project uses `github.com/dgrijalva/jwt-go`, which is no longer maintained. The community has forked it to `github.com/golang-jwt/jwt`.
  - **Impact**: The project is not receiving security patches for the JWT library, potentially exposing it to known vulnerabilities.
  - **Recommendation**: Migrate to a maintained JWT library, such as `github.com/golang-jwt/jwt/v4`.
- **[Code Smell] Global State in `init()`**
  - **Location**: `internal/service/user.go:102-105`
  - **Severity**: Major
  - **Description**: The `internal/service/user.go` file uses an `init()` function to initialize global repository variables. This creates tight coupling and makes testing difficult.
  - **Impact**: Code is difficult to unit test, and hidden dependencies make the system harder to reason about.
  - **Recommendation**: Use dependency injection to provide repositories to services. The repositories should be created in a central location (e.g., `main.go`) and passed to the services that need them.
- **[Code Smell] Global River Client**
    - **Location**: `internal/service/river.go:39`
    - **Severity**: Major
    - **Description**: The `GetRiverClient` function creates and configures a new River client on every call. This is inefficient and can lead to problems with connection pooling and resource management.
    - **Impact**: Inefficient resource usage, potential for connection leaks, and race conditions.
    - **Recommendation**: Create a single River client and share it across the application using dependency injection. Use a `sync.Once` to ensure the client is initialized only once.

## Minor Issues (Address When Possible)
- **[Security] Hardcoded JWT Expiration**
  - **Location**: `internal/service/user.go:26`
  - **Severity**: Minor
  - **Description**: The JWT expiration is hardcoded to one year, which is a very long time for a token to be valid.
  - **Impact**: A compromised token can be used for a long time, increasing the window of opportunity for an attacker.
  - **Recommendation**: Make the JWT expiration configurable via environment variables and consider using a shorter expiration time with refresh tokens.
- **[Code Quality] High Cyclomatic Complexity**
    - **Location**: `internal/service/definition_fetcher_worker.go:44`, `internal/service/definition_fetcher_worker.go:211`, `internal/service/leitner_system_strategy.go:526`
    - **Severity**: Minor
    - **Description**: Several functions have a high cyclomatic complexity, making them difficult to understand, test, and maintain.
    - **Impact**: Increased risk of bugs and difficulty in modifying the code.
    - **Recommendation**: Refactor these functions into smaller, more focused functions.
- **[Code Quality] Inconsistent Naming Conventions**
    - **Location**: Multiple files (see `golangci-lint` output)
    - **Severity**: Minor
    - **Description**: The codebase does not consistently follow Go's naming conventions (e.g., `Id` vs. `ID`, `Url` vs. `URL`).
    - **Impact**: Makes the code harder to read and understand.
    - **Recommendation**: Run `golangci-lint` with the `revive` linter enabled and fix the naming convention issues.
- **[Error Handling] Generic Error Messages**
  - **Location**: `internal/service/user.go:131`
  - **Severity**: Minor
  - **Description**: The `LoginUser` function returns a generic error message ("invalid combination of email and/or password"), which makes debugging difficult.
  - **Impact**: Difficult to diagnose the root cause of login failures.
  - **Recommendation**: Wrap errors with more context to provide more informative error messages. For example, use `fmt.Errorf("LoginUser: %w", err)`.

## Improvement Opportunities
- **Refactor to a Hexagonal Architecture**: The current 3-tier architecture is a good start, but a hexagonal architecture would provide better separation of concerns and make the application more modular and testable.
- **Implement a More Robust Configuration Management System**: The current system relies on environment variables, which can be difficult to manage in a complex application. Consider using a library like Viper to manage configuration from multiple sources (e.g., files, environment variables, remote key-value stores).
- **Introduce a Caching Layer**: For frequently accessed data that doesn't change often, introducing a caching layer (e.g., with Redis) could significantly improve performance and reduce database load.

## Metrics
- **Files reviewed**: 15
- **Issues found**: 12 (Critical: 0, Major: 4, Minor: 8)
- **Security vulnerabilities**: 3
- **Performance issues**: 0
- **Code quality issues**: 9

# Architecture Decision Records (ADRs)

## ADR-001: Migrate JWT Signing Algorithm to RS256

- **Decision**: Migrate the JWT signing algorithm from HS256 to RS256.
- **Status**: Accepted
- **Context**: The current implementation uses HS256, a symmetric algorithm, for signing JWTs. This is a security risk because if the secret key is compromised, an attacker can forge tokens. RS256 is an asymmetric algorithm that uses a public/private key pair, which is more secure.
- **Consequences**:
    - **Positive**: Improved security, as the private key can be kept secret on the server, while the public key can be distributed to clients for verification.
    - **Negative**: Increased complexity in key management.
- **Alternatives**:
    - **Stick with HS256**: This is not a viable option due to the security risks.
    - **Use another asymmetric algorithm (e.g., ES256)**: RS256 is a good choice because it is widely supported and well-understood.