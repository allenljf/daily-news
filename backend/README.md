# Daily News backend

FastAPI service for the single-user Daily News App.

## Local development

```sh
uv run ruff check .
uv run pytest -q
uv run uvicorn app.main:app --reload
```

Copy `.env.example` to `.env` and replace placeholder values only in your local environment. Production secrets are supplied through Cloud Run and Secret Manager.
