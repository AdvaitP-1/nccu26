"""Normalized symbol model for cross-language structural analysis.

Every language parser emits Symbol instances so that overlap detection
operates on a single, language-agnostic representation.
"""

from __future__ import annotations

from dataclasses import dataclass
from enum import Enum


class SymbolKind(str, Enum):
    FUNCTION = "function"
    ASYNC_FUNCTION = "async_function"
    CLASS = "class"
    METHOD = "method"
    IMPORT = "import"
    VARIABLE = "variable"
    CONSTANT = "constant"


@dataclass(frozen=True)
class Symbol:
    """A single code symbol extracted from source, normalized across languages."""

    name: str
    kind: SymbolKind
    start_line: int
    end_line: int
    file_path: str = ""
    parent: str | None = None

    @property
    def qualified_name(self) -> str:
        """Dot-separated name including parent scope, if any."""
        if self.parent:
            return f"{self.parent}.{self.name}"
        return self.name

    def overlaps_lines(self, other: Symbol) -> bool:
        """True when this symbol's line range intersects *other*'s."""
        return self.start_line <= other.end_line and other.start_line <= self.end_line
