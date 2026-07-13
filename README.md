<h1 align="center">GraphQL Microservice Architecture</h1>

<p align="center">
  A production-oriented e-commerce backend built with <strong>Go</strong>, demonstrating a robust <strong>Microservices Architecture</strong>. The system employs a <strong>GraphQL API Gateway</strong> as its single public entry point, delegating complex operations to highly decoupled internal microservices communicating over ultra-fast <strong>gRPC</strong> using <strong>Protocol Buffers</strong>.
</p>

<p align="center">
  This architecture was deliberately chosen to combine the flexibility of GraphQL (solving over-fetching and under-fetching for client applications) with the raw performance and strict contract definitions of gRPC for internal service-to-service communication.
</p>

---

## 🛠️ Technology Stack

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![GraphQL](https://img.shields.io/badge/GraphQL-E10098?style=for-the-badge&logo=graphql&logoColor=white)
![gRPC](https://img.shields.io/badge/gRPC-244C5A?style=for-the-badge&logo=grpc&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-316192?style=for-the-badge&logo=postgresql&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)
![Protocol Buffers](https://img.shields.io/badge/Protobuf-4285F4?style=for-the-badge&logo=google&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)

| Category | Technology |
|:---|:---|
| **Language** | Go (Golang 1.24+) |
| **API Gateway** | GraphQL (`gqlgen`) |
| **RPC / Internal Comms** | gRPC |
| **Data Serialization** | Protocol Buffers (Protobuf) |
| **Database** | PostgreSQL |
| **Containerization** | Docker & Docker Compose |
| **Workspace Management** | Go Workspaces (`go.work`) |

---

## ✨ Key Features

* **GraphQL API Gateway**: A single, flexible interface for clients resolving queries by aggregating data from internal services.
* **Microservices Architecture**: Completely isolated domain services (User, Order, Tracking) preventing monolithic entanglement.
* **gRPC Communication**: Lightning-fast, binary-encoded internal service communication.
* **Protocol Buffers**: Strongly typed API contracts acting as the single source of truth (`pkg/pb`).
* **PostgreSQL Integration**: Relational database persistence utilizing optimized connection pooling.
* **Dockerized Deployment**: Secure, non-root multi-stage Docker builds for minimal production image footprints.
* **Production-Ready Configuration**: Dynamic configuration utilizing `.env` files for seamless orchestration platforms (Render, Kubernetes).
* **Resiliency & Graceful Shutdown**: HTTP/gRPC server lifecycle management preventing hanging connections and dropped requests during deployments.
* **Health Probes**: Standardized HTTP (`/healthz`) and gRPC health checks configured for orchestrator monitoring.

---

## 🏗️ Architecture Overview

The system strictly encapsulates its logic. External clients communicate solely with the GraphQL Gateway. The Gateway parses the queries and acts as a gRPC client, forwarding requests to the respective backend services. The backend services then interact with the PostgreSQL database to persist and retrieve state.

```mermaid
graph TD
    Client((Client))

    subgraph "Public Facing"
        Gateway[GraphQL API Gateway<br>:8080]
    end

    subgraph "Internal gRPC Microservices"
        User[User Service<br>:50051]
        Order[Order Service<br>:50052]
        Tracking[Tracking Service<br>:50053]
    end

    DB[(PostgreSQL Database<br>:5432)]

    Client -- HTTP POST /query --> Gateway
    Gateway -- gRPC --> User
    Gateway -- gRPC --> Order
    Gateway -- gRPC --> Tracking

    User -- TCP --> DB
    Order -- TCP --> DB
    Tracking -- TCP --> DB
```

### Request Flow Example
1. A client requests an Order along with its Tracking status via GraphQL.
2. The Gateway receives the query.
3. The Gateway makes a gRPC call to the **Order Service** to fetch the order details.
4. The Gateway concurrently makes a gRPC call to the **Tracking Service** to fetch the tracking updates.
5. The Gateway aggregates the responses and returns a single JSON object back to the client.

---

## 📂 Repository Structure

The repository utilizes Go Workspaces (`go.work`) to manage multiple independent modules while sharing a centralized Protobuf module.

```text
.
├── api/
│   └── proto/                     # Single source of truth for Protobuf definitions (.proto)
├── pkg/
│   └── pb/                        # Shared Go Workspace module containing generated gRPC code
│       ├── orderpb/
│       ├── trackingpb/
│       └── userpb/
├── api_gateway/                   # GraphQL API Gateway (Independent Go Module)
│   ├── cmd/server/main.go         # Gateway entry point
│   ├── internal/graph/            # GraphQL Schema & generated code
│   └── Dockerfile                 
├── services/
│   ├── order_service/             # Order Domain Service (Independent Go Module)
│   │   ├── cmd/server/main.go
│   │   ├── internal/grpc_server/
│   │   └── Dockerfile
│   ├── tracking_service/          # Tracking Domain Service (Independent Go Module)
│   └── user_service/              # User Domain Service (Independent Go Module)
├── .env.example                   # Environment variables template
├── docker-compose.yml             # Local orchestration configuration
└── go.work                        # Go workspace defining local module relationships
```

---

## 🚀 Getting Started

### Prerequisites
* [Docker & Docker Compose](https://www.docker.com/get-started)
* [Go 1.24+](https://golang.org/dl/) (If running without Docker)
* [Make](https://www.gnu.org/software/make/) (Optional, for scripts)

### 1. Clone the Repository
```bash
git clone https://github.com/Shreyas071845/Microservices.git
cd Microservices
```

### 2. Configure Environment Variables
Copy the example environment file:
```bash
cp .env.example .env
```

### 3. Start Services (Docker Compose)
Docker Compose will build the gateway, the 3 microservices, and spin up a PostgreSQL instance.
```bash
docker-compose up --build
```
*(The GraphQL Gateway is now accessible at `http://localhost:8080/`)*

### 4. Stopping Services & Cleaning Up
```bash
# Stop and remove containers and networks
docker-compose down

# Stop, remove containers, networks, AND volumes (wipes the database)
docker-compose down -v
```

---

## ⚙️ Environment Variables

The system is fully configurable. The following variables are sourced from `.env`:

| Variable | Description | Default Value |
|:---|:---|:---|
| `PORT` | API Gateway HTTP port | `8080` |
| `USER_SERVICE_ADDR` | Address for gRPC User Service | `user-server:50051` |
| `ORDER_SERVICE_ADDR` | Address for gRPC Order Service | `order-server:50052` |
| `TRACKING_SERVICE_ADDR` | Address for gRPC Tracking Service | `tracking-server:50053` |
| `USER_SERVICE_PORT` | Port for the User Service to listen on | `50051` |
| `ORDER_SERVICE_PORT` | Port for the Order Service to listen on | `50052` |
| `TRACKING_SERVICE_PORT` | Port for the Tracking Service to listen on | `50053` |
| `DATABASE_URL` | PostgreSQL connection string | `postgres://devuser:devpassword@postgres-db:5432/assignment_db?sslmode=disable` |

---

## 💻 Running Without Docker

Thanks to `go.work`, running the microservices locally on bare-metal is seamless. Ensure you have a local PostgreSQL instance running and matching your `.env` `DATABASE_URL`.

Start each service in a separate terminal from the repository root:

```bash
# Terminal 1: User Service
cd services/user_service && go run ./cmd/server

# Terminal 2: Order Service
cd services/order_service && go run ./cmd/server

# Terminal 3: Tracking Service
cd services/tracking_service && go run ./cmd/server

# Terminal 4: API Gateway
cd api_gateway && go run ./cmd/server
```

---

## 📡 API Usage (GraphQL)

Access the GraphQL endpoint via POST requests to `http://localhost:8080/query`.

### 1. Create a User (Mutation)
**Request:**
```graphql
mutation {
  createUser(input: {
    name: "Jane Doe"
    email: "jane@example.com"
  }) {
    id
    name
    email
  }
}
```

**Response:**
```json
{
  "data": {
    "createUser": {
      "id": "a1b2c3d4",
      "name": "Jane Doe",
      "email": "jane@example.com"
    }
  }
}
```

### 2. Create an Order (Mutation)
**Request:**
```graphql
mutation {
  createOrder(input: {
    userId: "a1b2c3d4"
    productName: "Mechanical Keyboard"
    quantity: 1
  }) {
    id
    userId
    productName
    status
  }
}
```

**Response:**
```json
{
  "data": {
    "createOrder": {
      "id": "x9y8z7",
      "userId": "a1b2c3d4",
      "productName": "Mechanical Keyboard",
      "status": "PENDING"
    }
  }
}
```

### 3. Query All Users (Query)
**Request:**
```graphql
query {
  listUsers {
    id
    name
    email
  }
}
```

**Response:**
```json
{
  "data": {
    "listUsers": [
      {
        "id": "a1b2c3d4",
        "name": "Jane Doe",
        "email": "jane@example.com"
      }
    ]
  }
}
```

---

## 🧩 Service Details

| Service | Exposes | Responsibilities | Protocol |
|:---|:---|:---|:---|
| **API Gateway** | `:8080` (HTTP) | Single entry point. Aggregates data via GraphQL resolvers. Maps GraphQL to gRPC. | HTTP/GraphQL |
| **User Service** | `:50051` (gRPC) | Manages User lifecycle, persistence, and queries. | gRPC |
| **Order Service** | `:50052` (gRPC) | Manages Order lifecycle. Validates relationships (e.g., User exists). | gRPC |
| **Tracking Service**| `:50053` (gRPC) | Manages shipping updates, geo-location state, and tracking statuses. | gRPC |

---

## 🐳 Docker Architecture

The project emphasizes immutable, production-ready containers.

* **Multi-stage Builds**: Every `Dockerfile` compiles the Go binary using `golang:alpine` and copies it into a lightweight, scratch/alpine image.
* **Security**: Binaries are executed by an unprivileged `appuser` user rather than `root`.
* **Workspace Resolution**: Docker build contexts are set to the root directory (`.`) to allow the compiler to pull the shared `pkg/pb` module effortlessly.
* **Internal Networking**: Services communicate over Docker Compose's bridge network using internal DNS resolution (e.g., `user-server:50051`).

---

## 🛡️ Production Readiness

This project implements rigorous production-hardening standards:

* **Graceful Shutdown**: The API Gateway and gRPC servers intercept `SIGINT`/`SIGTERM` to safely drain active requests before terminating.
* **Connection Pooling**: PostgreSQL `*sql.DB` instances enforce explicit limits (`SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime`) preventing port exhaustion.
* **gRPC Health Probes**: Backend services expose standard `grpc_health_v1` endpoints for Kubernetes/Cloud Orchestrator liveness/readiness probes.
* **HTTP Server Timeouts**: The Gateway's `http.Server` utilizes strict `ReadTimeout`, `WriteTimeout`, and `IdleTimeout` bounds to mitigate slowloris attacks.
* **gRPC Retries**: Inter-service communication benefits from transparent, policy-driven gRPC retries.

---

## 📈 Future Improvements

While this is a robust foundation, a highly-scaled enterprise system would further introduce:

* **Authentication & Authorization**: Implement JWT validation at the API Gateway and pass claims via gRPC metadata.
* **Caching**: Introduce a `Redis` layer in front of PostgreSQL for heavy read-path queries (like `listUsers`).
* **Distributed Tracing**: Integrate OpenTelemetry (`Jaeger`/`Zipkin`) to trace requests as they jump between the Gateway and gRPC boundaries.
* **Message Broker (Event-Driven)**: Introduce `Kafka` or `RabbitMQ`. (e.g., Creating an order publishes an `OrderCreated` event that the Tracking Service consumes asynchronously instead of blocking).
* **Metrics**: Expose Prometheus metrics endpoints to monitor GraphQL latency and gRPC error rates via Grafana.

---

## 📸 Screenshots

*(Placeholders for project demonstrations)*

| GraphQL Playground | Docker Compose Logs |
|:---:|:---:|
| `[GraphQL Screenshot Placeholder]` | `[Docker Logs Screenshot Placeholder]` |

| Postman Queries | Terminal Outputs |
|:---:|:---:|
| `[Postman Screenshot Placeholder]` | `[Terminal Logs Screenshot Placeholder]` |

---

## 🌐 Live Deployment

* **Live GraphQL API:** `[Render/Fly.io Placeholder]`
* **Health Endpoint:** `[Health Check URL Placeholder]`
* **GitHub Repository:** `[GitHub URL Placeholder]`

---

## 🧠 Design Decisions: Why this Stack?

* **Why a GraphQL Gateway?** In an E-commerce domain, clients often need deeply nested data (e.g., A user, their orders, and the tracking status of each order). A REST API would either require multiple round-trips or a massive over-fetching monolithic endpoint. GraphQL solves this elegantly.
* **Why gRPC internally?** While HTTP/JSON is great for public APIs, it's inefficient for internal microservices. gRPC utilizes HTTP/2 and binary Protobuf serialization, making it incredibly fast and strongly typed, ensuring breaking changes are caught at compile time.
* **Why Go?** Go was built for network services. Its lightweight goroutines naturally handle high-concurrency RPC connections, and its fast compile times and statically linked binaries make Docker deployments trivial and extremely small.
* **Why a Shared Protobuf Module?** Eliminates code duplication. All microservices depend on a single source of truth for their data contracts, drastically reducing the chances of structural drift.

---

## 🤝 Contributing

Contributions are what make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

---

## 📄 License

Copyright (c) 2026 Shreyas.

Distributed under the MIT License. See `LICENSE` for more information.