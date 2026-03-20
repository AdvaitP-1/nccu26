"""Centralised configuration for the agent swarm demo."""

import os

MCP_SSE_URL = os.environ.get("MCP_SSE_URL", "http://localhost:9090")
MCP_HTTP_URL = os.environ.get("MCP_HTTP_URL", "http://localhost:9091")
BACKEND_URL = os.environ.get("BACKEND_URL", "http://localhost:8000")

DEMO_BRANCH = "feature/swarm-demo"
