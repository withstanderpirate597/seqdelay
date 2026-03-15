---
sidebar_position: 5
---

# HTTP API

## Starting the Server

```go
srv := seqdelay.NewServer(q, seqdelay.WithServerAddr(":9280"))
srv.ListenAndServe()
```

## Endpoints

### POST /add

Add a delayed task.

```json
{
  "topic": "order-cancel",
  "id": "order-123",
  "body": "{\"order_id\":\"123\"}",
  "delay_ms": 900000,
  "ttr_ms": 30000,
  "max_retries": 3
}
```

**Response:** `200` on success, `409` on duplicate.

### POST /pop

Pull a ready task. Long-polls until a task is available.

```json
{
  "topic": "order-cancel",
  "timeout_ms": 30000
}
```

**Response:** `200` with task data, or `200` with `null` data on timeout.

### POST /finish

Mark a task as completed.

```json
{
  "topic": "order-cancel",
  "id": "order-123"
}
```

**Response:** `200` on success, `404` if not found, `400` if not in ACTIVE state.

### POST /cancel

Cancel a pending task.

```json
{
  "topic": "order-cancel",
  "id": "order-123"
}
```

**Response:** `200` on success, `404` if not found.

### GET /get

Query task details.

```
GET /get?topic=order-cancel&id=order-123
```

**Response:** `200` with full task object.

### GET /stats

Queue statistics.

```
GET /stats
```

**Response:**

```json
{
  "code": 0,
  "data": {
    "topics": {
      "order-cancel": { "ready": 10, "active": 3 },
      "payment-notify": { "ready": 0, "active": 1 }
    }
  }
}
```

## Response Format

```json
{
  "code": 0,
  "message": "ok",
  "data": null
}
```

| code | Meaning |
|------|---------|
| 0 | Success |
| non-0 | Error (see message) |

## HTTP Status Codes

| Status | When |
|--------|------|
| 200 | Success |
| 400 | Validation error, invalid JSON |
| 404 | Task not found |
| 409 | Duplicate task or topic conflict |
| 500 | Internal error |
| 503 | Queue is closed |
