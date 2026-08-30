from fastapi import FastAPI

from app.categories.router import router as categories_router
from app.core.config import get_settings


def create_app() -> FastAPI:
    """Create the Daily News HTTP application."""
    get_settings()
    app = FastAPI(title="Daily News API")
    app.include_router(categories_router)

    @app.get("/healthz")
    async def healthz() -> dict[str, str]:
        return {"status": "ok"}

    return app


app = create_app()
