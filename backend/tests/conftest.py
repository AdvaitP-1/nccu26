"""Shared test fixtures and helpers for the backend test suite."""

from __future__ import annotations

import pytest
from fastapi.testclient import TestClient

from app.main import app
from app.schemas import AnalyzeOverlapsRequest, ChangeSet, FileInput


@pytest.fixture()
def client() -> TestClient:
    return TestClient(app)


# ---------------------------------------------------------------------------
# Helpers for building requests
# ---------------------------------------------------------------------------

def make_changeset(agent_id: str, files: list[tuple[str, str, str]]) -> ChangeSet:
    """Shorthand: files is a list of (path, language, content) tuples."""
    return ChangeSet(
        agent_id=agent_id,
        files=[FileInput(path=p, language=l, content=c) for p, l, c in files],
    )


def make_request(*changesets: ChangeSet) -> AnalyzeOverlapsRequest:
    return AnalyzeOverlapsRequest(changesets=list(changesets))


# Common source snippets
PY_FUNC_A = """\
def validate_token(token):
    return True

def refresh_token(t):
    pass
"""

PY_FUNC_B = """\
def validate_token(tok):
    if not tok:
        raise ValueError
    return check(tok)

class AuthManager:
    pass
"""

PY_FUNC_C = """\
def validate_token(t):
    return verify(t)
"""

PY_SEPARATE = """\
def unrelated_function():
    return 42
"""

PY_INVALID = """\
def broken(
    # missing close paren and colon
"""

TS_FUNC_A = """\
export function fetchData(url: string) {
  return fetch(url);
}
"""

TS_FUNC_B = """\
export function fetchData(endpoint: string) {
  const res = await fetch(endpoint);
  return res.json();
}
"""

PY_CLASS_METHODS = """\
class Service:
    def start(self):
        pass

    def stop(self):
        pass
"""

PY_CLASS_METHODS_ALT = """\
class Service:
    def start(self):
        self.running = True

    def health(self):
        return True
"""

PY_IMPORTS_ONLY = """\
import os
from pathlib import Path
"""

PY_TOPLEVEL_VAR = """\
MAX_RETRIES = 5
timeout = 30
"""
