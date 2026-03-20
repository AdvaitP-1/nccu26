"""FastAPI application entry-point for the structural analysis backend.

Run locally:
    uvicorn app.main:app --reload --port 8000

This service is an *internal* dependency of the MCP layer.  It should never
be exposed directly to agents or external orchestrators.
"""

from __future__ import annotations

import logging

from fastapi import FastAPI

from app.api.routes import router
from app.api.tree_routes import router as tree_router
from app.config import settings

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s  %(levelname)-8s  %(name)s  %(message)s",
)

app = FastAPI(
    title="Structural Analysis Backend",
    version="0.1.0",
    docs_url="/docs",
    redoc_url=None,
)

app.include_router(router)
app.include_router(tree_router)


@app.on_event("startup")
async def _log_config() -> None:
    logging.getLogger(__name__).info(
        "Backend starting — host=%s port=%s", settings.host, settings.port
    )
