from fastapi import FastAPI

from app.categories.router import router as categories_router
from app.core.config import get_settings
from app.ingestion.router import router as ingestion_runs_router
from app.news.router import router as news_router


def create_app() -> FastAPI:
    """Create the Daily News HTTP application."""
    get_settings()
    app = FastAPI(title="Daily News API")
    app.include_router(categories_router)
    app.include_router(ingestion_runs_router)
    app.include_router(news_router)

    @app.get("/healthz")
    async def healthz() -> dict[str, str]:
        return {"status": "ok"}

    return app


app = create_app()
