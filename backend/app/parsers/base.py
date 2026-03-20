"""Abstract interface that every language parser must implement."""

from __future__ import annotations

from abc import ABC, abstractmethod

from app.models.symbols import Symbol


class BaseParser(ABC):
    """Language-agnostic parser contract.

    Implementations must:
      1. Return a list of ``Symbol`` instances on success.
      2. Never raise on malformed input — return ``[]`` and log the issue.
    """

    @abstractmethod
    def parse(self, source: str, file_path: str) -> list[Symbol]:
        """Extract normalised symbols from *source*."""
        ...

    @abstractmethod
    def supported_extensions(self) -> list[str]:
        """File extensions this parser can handle (e.g. ``['.py']``)."""
        ...
