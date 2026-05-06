# ahri-health-bridge

FastAPI receiver for Apple Health data from iOS Shortcuts to Neon (PostgreSQL).

## Requirements
- FastAPI
- Uvicorn
- Pydantic
- Psycopg2-binary
- Python-dotenv

## Models
- **Steps**: count, source, timestamp.
- **Sleep**: start_time, end_time, duration, type.
- **Heart Rate**: bpm, timestamp.

## Security
- `X-API-Key` header authentication required for all ingestion endpoints.
