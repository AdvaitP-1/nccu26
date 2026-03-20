"""Go parser using regex-based extraction.

Extracts functions, methods, structs, interfaces, type declarations,
imports, constants, and variables from Go source code.  Uses the stdlib
``re`` module rather than tree-sitter to keep dependencies minimal.
"""

from __future__ import annotations

import logging
import re

from app.models.symbols import Symbol, SymbolKind
from app.parsers.base import BaseParser

logger = logging.getLogger(__name__)

# Patterns are ordered from most specific to least.
_PATTERNS: list[tuple[re.Pattern[str], SymbolKind, int]] = [
    # Method with receiver: func (r *Receiver) Name(...)
    (re.compile(r"^func\s+\([^)]+\)\s+(\w+)\s*\("), SymbolKind.METHOD, 1),
    # Standalone function: func Name(...)
    (re.compile(r"^func\s+(\w+)\s*\("), SymbolKind.FUNCTION, 1),
    # Struct: type Name struct {
    (re.compile(r"^type\s+(\w+)\s+struct\b"), SymbolKind.CLASS, 1),
    # Interface: type Name interface {
    (re.compile(r"^type\s+(\w+)\s+interface\b"), SymbolKind.CLASS, 1),
    # Other type alias: type Name = ... / type Name ...
    (re.compile(r"^type\s+(\w+)\b"), SymbolKind.CLASS, 1),
    # Import single: import "pkg"
    (re.compile(r'^import\s+"([^"]+)"'), SymbolKind.IMPORT, 1),
    # Const block start: const ( or const Name = ...
    (re.compile(r"^const\s+(\w+)"), SymbolKind.CONSTANT, 1),
    # Var block start: var ( or var Name ...
    (re.compile(r"^var\s+(\w+)"), SymbolKind.VARIABLE, 1),
]


class GoParser(BaseParser):
    def supported_extensions(self) -> list[str]:
        return [".go"]

    def parse(self, source: str, file_path: str) -> list[Symbol]:
        symbols: list[Symbol] = []
        lines = source.split("\n")

        # Track multi-line constructs
        in_func: str | None = None
        func_start = 0
        brace_depth = 0

        for lineno_0, line in enumerate(lines):
            lineno = lineno_0 + 1
            stripped = line.strip()

            # Track brace depth for end-of-function detection
            brace_depth += stripped.count("{") - stripped.count("}")

            # If we're inside a function, check if it ended
            if in_func is not None and brace_depth <= 0:
                # Update the last symbol's end_line
                for i in range(len(symbols) - 1, -1, -1):
                    if symbols[i].name == in_func and symbols[i].start_line == func_start:
                        symbols[i] = Symbol(
                            name=symbols[i].name,
                            kind=symbols[i].kind,
                            start_line=symbols[i].start_line,
                            end_line=lineno,
                            file_path=file_path,
                            parent=symbols[i].parent,
                        )
                        break
                in_func = None
                func_start = 0

            # Try each pattern on the stripped line
            for pattern, kind, group in _PATTERNS:
                m = pattern.search(stripped)
                if m:
                    name = m.group(group)

                    # Determine parent for methods
                    parent: str | None = None
                    if kind == SymbolKind.METHOD:
                        recv_match = re.search(r"\(\w+\s+\*?(\w+)\)", stripped)
                        if recv_match:
                            parent = recv_match.group(1)

                    symbols.append(
                        Symbol(
                            name=name,
                            kind=kind,
                            start_line=lineno,
                            end_line=lineno,
                            file_path=file_path,
                            parent=parent,
                        )
                    )

                    # Track function/method bodies for accurate end_line
                    if kind in (SymbolKind.FUNCTION, SymbolKind.METHOD):
                        in_func = name
                        func_start = lineno
                        brace_depth = stripped.count("{") - stripped.count("}")

                    break

            # Handle import blocks: import ( ... )
            if stripped.startswith("import ("):
                continue
            if stripped.startswith('"') and stripped.endswith('"'):
                # Inside an import block
                pkg = stripped.strip('"')
                if "/" in pkg:
                    pkg = pkg.rsplit("/", 1)[-1]
                symbols.append(
                    Symbol(
                        name=pkg,
                        kind=SymbolKind.IMPORT,
                        start_line=lineno,
                        end_line=lineno,
                        file_path=file_path,
                    )
                )

        return symbols
