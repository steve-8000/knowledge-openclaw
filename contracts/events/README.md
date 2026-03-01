# Event Contracts

Canonical event types for the ki-db indexing pipeline.
All events are published via NATS JetStream on `idx.doc.{EventType}`.

## Event Types

| Type | Subject | Producer | Consumer(s) |
|------|---------|----------|-------------|
| `DocumentUpserted` | `idx.doc.DocumentUpserted` | Ingest API | - |
| `VersionCreated` | `idx.doc.VersionCreated` | Ingest API | Parser |
| `DocumentParsed` | `idx.doc.DocumentParsed` | Parser | Chunker, Graph |
| `ChunksCreated` | `idx.doc.ChunksCreated` | Chunker | Embedder |
| `EmbeddingsGenerated` | `idx.doc.EmbeddingsGenerated` | Embedder | - |
| `GraphUpdated` | `idx.doc.GraphUpdated` | Graph Builder | - |
| `IndexJobFailed` | `idx.doc.IndexJobFailed` | Any Worker | Quality |
| `FeedbackReceived` | `idx.doc.FeedbackReceived` | Query API | Quality |

## Pipeline Flow

```
Ingest → DocumentUpserted + VersionCreated
           ↓
        Parser → DocumentParsed
           ↓
        Chunker → ChunksCreated
           ↓
        Embedder → EmbeddingsGenerated
           ↓
        Graph Builder → GraphUpdated
```

## Envelope Format

Every event is wrapped in an envelope:

```json
{
  "event_id": "uuid",
  "event_type": "VersionCreated",
  "tenant_id": "uuid",
  "timestamp": "2024-01-01T00:00:00Z",
  "payload": { ... }
}
```

See `backend/pkg/events/events.go` for Go type definitions.
