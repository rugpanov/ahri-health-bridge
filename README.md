# ahri-health-bridge

HTTP receiver for Apple Health data from iOS Shortcuts to Neon (PostgreSQL), written in Go.

## Requirements
- Go 1.22+
- Neon (PostgreSQL) database

## Models
- **Steps**: count, source, timestamp

Designed to be extended with additional health parameters (sleep, heart rate, etc.).

## Security
- `X-API-Key` header authentication required for all ingestion endpoints.
