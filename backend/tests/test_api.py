"""Integration tests hitting the FastAPI endpoints end-to-end."""

from __future__ import annotations

import pytest
from fastapi.testclient import TestClient

from app.main import app
from tests.conftest import PY_FUNC_A, PY_FUNC_B, PY_SEPARATE


@pytest.fixture()
def client():
    return TestClient(app)


class TestHealthEndpoint:
    def test_health_ok(self, client: TestClient):
        resp = client.get("/health")
        assert resp.status_code == 200
        assert resp.json()["status"] == "ok"


class TestAnalyzeOverlapsEndpoint:
    def test_empty_changesets_returns_empty(self, client: TestClient):
        resp = client.post("/analyze/overlaps", json={"changesets": []})
        assert resp.status_code == 200
        body = resp.json()
        assert body["overlaps"] == []
        assert body["file_risks"] == []

    def test_two_agents_overlapping(self, client: TestClient):
        resp = client.post("/analyze/overlaps", json={
            "changesets": [
                {"agent_id": "a", "files": [{"path": "f.py", "language": "python", "content": PY_FUNC_A}]},
                {"agent_id": "b", "files": [{"path": "f.py", "language": "python", "content": PY_FUNC_B}]},
            ]
        })
        assert resp.status_code == 200
        body = resp.json()
        assert len(body["overlaps"]) >= 1
        assert len(body["file_risks"]) >= 1
        fr = body["file_risks"][0]
        assert "contributors" in fr
        assert "contributors_count" in fr
        assert "max_severity" in fr
        assert "is_hotspot" in fr

    def test_no_overlap_different_symbols(self, client: TestClient):
        resp = client.post("/analyze/overlaps", json={
            "changesets": [
                {"agent_id": "a", "files": [{"path": "f.py", "language": "python", "content": PY_FUNC_A}]},
                {"agent_id": "b", "files": [{"path": "f.py", "language": "python", "content": PY_SEPARATE}]},
            ]
        })
        assert resp.status_code == 200
        body = resp.json()
        assert body["overlaps"] == []

    def test_five_agents_response_structure(self, client: TestClient):
        changesets = [
            {"agent_id": f"agent-{i}", "files": [
                {"path": "hot.py", "language": "python", "content": PY_FUNC_A}
            ]}
            for i in range(5)
        ]
        resp = client.post("/analyze/overlaps", json={"changesets": changesets})
        assert resp.status_code == 200
        body = resp.json()
        fr = body["file_risks"][0]
        assert fr["contributors_count"] == 5
        assert fr["is_hotspot"] is True
        # C(5,2) = 10 pairwise overlap records for validate_token alone
        assert fr["pairwise_overlap_count"] >= 10

    def test_unsupported_extension_no_crash(self, client: TestClient):
        resp = client.post("/analyze/overlaps", json={
            "changesets": [
                {"agent_id": "a", "files": [{"path": "f.go", "language": "go", "content": "package main"}]},
                {"agent_id": "b", "files": [{"path": "f.go", "language": "go", "content": "package main"}]},
            ]
        })
        assert resp.status_code == 200
        assert resp.json()["overlaps"] == []
