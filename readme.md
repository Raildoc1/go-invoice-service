# 🧾 Invoice Service

Example of a modern Go microservice for managing invoices, built with persistence, and message streaming.

---

## ✨ Features

- 📄 [RESTful API for creating and retrieving invoices](./services/api-service/internal/httpserver/server.go)
- 🛢️ [PostgreSQL for persistent storage](./services/storage-service/internal/data/postgres)
- 🔁 [Kafka integration using an Outbox pattern](./services/message-scheduler-service)
- 📊 [Prometheus metrics for performance and business KPIs](./common/pkg/meterutils/prometheus-server.go)
- 🧠 [gRPC endpoint for inter-service communication](./proto)
- 🧪 [Auto-generated mocks with `mockgen`](./services/validation-service/internal/services)
- 🐳 [Docker Compose for easy local development](./docker-compose.yaml)

---

## 📦 Tech Stack

| Layer        | Tech                       |
|--------------|----------------------------|
| Language     | Go (1.24)                  |
| Transport    | `chi` (REST), gRPC         |
| DB           | `sqlc` (PostgreSQL)        |
| Testing      | `mockgen`, `testify`       |
| Queue        | Kafka                      |
| Metrics      | OpenTelemetry + Prometheus |
| Build/Deploy | Docker, Docker Compose     |

## 🧩 System Architecture

This service uses a modular microservice pattern connected via HTTP/gRPC and Kafka.

```text
                  ┌────────────────────┐
                  │   API Service      │
                  │  (REST/gRPC)       │
                  └────────┬───────────┘
                           │
                  ┌────────▼───────────┐
    Postgres  <-  │  Storage Service   │  -> new_invoice Kafka message (saved to outbox)
                  └────────────────────┘

                  ┌────────────────────┐
      Outbox  ->  │ Message Scheduler  │  -> Kafka
                  └────────────────────┘

                  ┌────────────────────┐
                  │ Validation Service │  <- new_invoice Kafka message
                  └────────┬───────────┘
                           │
                  ┌────────▼───────────┐
    Postgres  <-  │ Storage Service    │  -> invoice_approved/invoice_rejected Kafka message (saved to outbox)
                  └────────────────────┘     that can be later consumed e.g. by notifications service
```

## 🚀 Getting Started

```bash
docker-compose up --build
```

## 🧪 API Service Endpoints

| Method | Path                  | Description          | Request Body     |
|--------|-----------------------|----------------------|------------------|
| `POST` | `/api/invoice/create` | Create a new invoice | JSON (see below) |
| `POST` | `/api/invoice/get`    | Get invoice by ID    | JSON (see below) |

## 📥 Example: Create Invoice Request

### Request

```http
POST /api/invoice/create
Content-Type: application/json
```

### Response

```json
{
  "invoice": {
    "id": "53150a25-02f1-540a-99e7-48e267fd6d13",
    "customer_id": "c78aef21-ae9f-4561-a2c9-3b7a7ea2f990",
    "amount": 1050.00,
    "currency": "USD",
    "due_date": "2025-06-30T00:00:00Z",
    "created_at": "2025-06-01T15:04:05Z",
    "updated_at": "2025-06-10T10:22:30Z",
    "items": [
      {
        "description": "Website Design",
        "quantity": 1,
        "unit_price": 1000.0,
        "total": 1000.0
      },
      {
        "description": "Hosting (1 month)",
        "quantity": 1,
        "unit_price": 50.0,
        "total": 50.0
      }
    ],
    "notes": "Payment due within 30 days."
  }
}

```

---

## 📤 Example: Get Invoice Response

### Request

```http
POST /api/invoice/get
Content-Type: application/json
```

```json
{
  "id": "53150a25-02f1-540a-99e7-48e267fd6d13"
}
```

### Response

```json
{
  "invoice": {
    "id": "53150a25-02f1-540a-99e7-48e267fd6d13",
    "customer_id": "c78aef21-ae9f-4561-a2c9-3b7a7ea2f990",
    "amount": "1050",
    "currency": "USD",
    "due_date": "2025-06-30T00:00:00Z",
    "created_at": "2025-06-01T15:04:05Z",
    "updated_at": "2025-06-10T10:22:30Z",
    "items": [
      {
        "description": "Website Design",
        "quantity": 1,
        "unit_price": "1000",
        "total": "1000"
      },
      {
        "description": "Hosting (1 month)",
        "quantity": 1,
        "unit_price": "50",
        "total": "50"
      }
    ],
    "notes": "Payment due within 30 days."
  },
  "status": "Approved"
}
```

---

## 📊 Metrics Exposed

Metrics available at `/metrics`:

### API Service

- `http_requests_total`
- `http_request_duration_seconds`

### Message schedule service

- `kafka_total_produce_messages`
- `kafka_total_produce_bytes`

### Validation service

- `kafka_total_consumed_messages`
- `total_handled_invoices`
