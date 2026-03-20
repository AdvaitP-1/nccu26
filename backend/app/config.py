"""Backend service configuration.

All tunables are read from environment variables at startup with sensible
defaults for local development.
"""

from __future__ import annotations

import os
from dataclasses import dataclass


@dataclass(frozen=True)
class Settings:
    host: str = "0.0.0.0"
    port: int = 8000

    # Risk-score weights per overlap severity.
    critical_weight: int = 40
    high_weight: int = 25
    medium_weight: int = 15
    low_weight: int = 5

    # A file with zero overlaps is assigned this stability baseline.
    base_stability: int = 100

    # Maximum risk score (used to clamp the final value).
    max_risk: int = 100

    @classmethod
    def from_env(cls) -> Settings:
        return cls(
            host=os.getenv("BACKEND_HOST", "0.0.0.0"),
            port=int(os.getenv("BACKEND_PORT", "8000")),
            critical_weight=int(os.getenv("BACKEND_CRITICAL_WEIGHT", "40")),
            high_weight=int(os.getenv("BACKEND_HIGH_WEIGHT", "25")),
            medium_weight=int(os.getenv("BACKEND_MEDIUM_WEIGHT", "15")),
            low_weight=int(os.getenv("BACKEND_LOW_WEIGHT", "5")),
            base_stability=int(os.getenv("BACKEND_BASE_STABILITY", "100")),
            max_risk=int(os.getenv("BACKEND_MAX_RISK", "100")),
        )


settings = Settings.from_env()
