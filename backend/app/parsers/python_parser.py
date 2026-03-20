"""Python parser using the built-in ``ast`` module."""

from __future__ import annotations

import ast
import logging

from app.models.symbols import Symbol, SymbolKind
from app.parsers.base import BaseParser

logger = logging.getLogger(__name__)


class PythonParser(BaseParser):
    def supported_extensions(self) -> list[str]:
        return [".py"]

    def parse(self, source: str, file_path: str) -> list[Symbol]:
        try:
            tree = ast.parse(source, filename=file_path)
        except SyntaxError as exc:
            logger.warning("Python parse failed for %s: %s", file_path, exc)
            return []

        symbols: list[Symbol] = []
        self._visit(tree, file_path, symbols, parent=None)
        return symbols

    # ------------------------------------------------------------------
    # Internal recursive visitor
    # ------------------------------------------------------------------

    def _visit(
        self,
        node: ast.AST,
        file_path: str,
        out: list[Symbol],
        parent: str | None,
    ) -> None:
        for child in ast.iter_child_nodes(node):
            match child:
                case ast.FunctionDef():
                    kind = SymbolKind.METHOD if parent else SymbolKind.FUNCTION
                    out.append(self._sym(child.name, kind, child, file_path, parent))
                    self._visit(child, file_path, out, parent=child.name)

                case ast.AsyncFunctionDef():
                    kind = SymbolKind.METHOD if parent else SymbolKind.ASYNC_FUNCTION
                    out.append(self._sym(child.name, kind, child, file_path, parent))
                    self._visit(child, file_path, out, parent=child.name)

                case ast.ClassDef():
                    out.append(
                        self._sym(child.name, SymbolKind.CLASS, child, file_path, parent)
                    )
                    self._visit(child, file_path, out, parent=child.name)

                case ast.Import():
                    for alias in child.names:
                        out.append(
                            self._sym(alias.name, SymbolKind.IMPORT, child, file_path, parent)
                        )

                case ast.ImportFrom():
                    module = child.module or ""
                    for alias in child.names:
                        out.append(
                            self._sym(
                                f"{module}.{alias.name}",
                                SymbolKind.IMPORT,
                                child,
                                file_path,
                                parent,
                            )
                        )

                case ast.Assign() if parent is None:
                    for target in child.targets:
                        if isinstance(target, ast.Name):
                            kind = (
                                SymbolKind.CONSTANT
                                if target.id.isupper()
                                else SymbolKind.VARIABLE
                            )
                            out.append(
                                self._sym(target.id, kind, child, file_path, parent)
                            )

                case ast.AnnAssign() if parent is None:
                    if isinstance(child.target, ast.Name):
                        out.append(
                            self._sym(
                                child.target.id,
                                SymbolKind.VARIABLE,
                                child,
                                file_path,
                                parent,
                            )
                        )

                case _:
                    self._visit(child, file_path, out, parent=parent)

    @staticmethod
    def _sym(
        name: str,
        kind: SymbolKind,
        node: ast.AST,
        file_path: str,
        parent: str | None,
    ) -> Symbol:
        return Symbol(
            name=name,
            kind=kind,
            start_line=getattr(node, "lineno", 0),
            end_line=getattr(node, "end_lineno", None) or getattr(node, "lineno", 0),
            file_path=file_path,
            parent=parent,
        )
