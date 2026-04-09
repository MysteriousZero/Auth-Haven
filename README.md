# Auth Haven

**Auth Haven** is a robust, spec-driven Authentication and Authorization service built with Go, gRPC, and REST.

The project follows a **Spec-Driven Development** approach, where implementation is continuously assessed against a set of formal specifications in the `spec/` directory.

## Current Progress

For the latest implementation score and compliance gaps, please refer to:
👉 **[spec_implementation_status.md](./spec_implementation_status.md)**

---

## 🛠️ Development Environment

The project includes a fully-configured **VS Code Dev Container** for a consistent development experience. It comes pre-installed with:
- **Go toolchain** and language server.
- **Protobuf/gRPC tools** (including UI helpers).
- **Linters** (`golangci-lint`).
- **Docker** support.

We **strongly recommend** using this setup if you are using VS Code. Simply open the folder and select **"Reopen in Container"** when prompted.

---

## Getting Started

### 📡 Protobuf / gRPC Generation

To generate the Go code from your `.proto` definitions, run the following command from the project root:

```bash
protoc --go_out=. \
       --go-grpc_out=. \
       --go_opt=paths=source_relative \
       --go-grpc_opt=paths=source_relative \
       api/proto/*.proto \
       api/proto/common/*.proto
```

Generated files will be located in `pkg/proto/`.

### 🐳 Running with Docker

The easiest way to start the application (including all dependencies) is via Docker Compose:

```bash
docker compose up --build
```

The application will be available at [http://localhost:8080](http://localhost:8080).

---

## Testing

For comprehensive test coverage status and production readiness assessment, please refer to:
**[tests.md](./tests.md)** - Complete test coverage analysis across all architectural layers

**Current Status**: Repository layer has basic CRUD coverage but **missing critical transaction and rollback testing**.

---

## Project Structure

*   `api/proto/`: Source Protobuf definitions.
*   `cmd/`: Main entry points for the server.
*   `internal/`: Private application code (Repositories, Services, Handlers).
*   `pkg/proto/`: Generated gRPC and Protobuf code.
*   `spec/`: Core specifications (DB, Domain, gRPC, REST, Security).
*   `test/`: Comprehensive test suite including repository and integration tests.
