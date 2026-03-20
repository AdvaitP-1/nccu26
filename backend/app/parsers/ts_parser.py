"""TypeScript / JavaScript / TSX / JSX parser using tree-sitter.

Requires:
    pip install tree-sitter tree-sitter-javascript tree-sitter-typescript

If the tree-sitter packages are missing at import time the parser will
still be importable but ``parse()`` will return an empty symbol list and
log a warning.
"""

from __future__ import annotations

import logging
from typing import TYPE_CHECKING

from app.models.symbols import Symbol, SymbolKind
from app.parsers.base import BaseParser

if TYPE_CHECKING:
    from tree_sitter import Node

logger = logging.getLogger(__name__)

# ---------------------------------------------------------------------------
# Lazy language loading — fail gracefully when packages are absent.
# ---------------------------------------------------------------------------

_TREE_SITTER_AVAILABLE = False

try:
    import tree_sitter_javascript as _tsjs
    import tree_sitter_typescript as _tsts
    from tree_sitter import Language, Parser

    _JS_LANG = Language(_tsjs.language())
    _TS_LANG = Language(_tsts.language_typescript())
    _TSX_LANG = Language(_tsts.language_tsx())
    _TREE_SITTER_AVAILABLE = True
except Exception:  # noqa: BLE001 — ImportError *or* ABI mismatch
    logger.warning(
        "tree-sitter packages not installed or incompatible; TS/JS parsing disabled"
    )

_EXT_TO_LANG: dict[str, object] = {}
if _TREE_SITTER_AVAILABLE:
    _EXT_TO_LANG = {
        ".ts": _TS_LANG,
        ".tsx": _TSX_LANG,
        ".js": _JS_LANG,
        ".jsx": _JS_LANG,
    }


# ---------------------------------------------------------------------------
# Parser implementation
# ---------------------------------------------------------------------------


class TypeScriptParser(BaseParser):
    def supported_extensions(self) -> list[str]:
        return [".ts", ".tsx", ".js", ".jsx"]

    def parse(self, source: str, file_path: str) -> list[Symbol]:
        if not _TREE_SITTER_AVAILABLE:
            logger.error("Cannot parse %s — tree-sitter unavailable", file_path)
            return []

        ext = _ext_of(file_path)
        lang = _EXT_TO_LANG.get(ext)
        if lang is None:
            logger.warning("No language grammar for extension %s", ext)
            return []

        parser = Parser(lang)
        try:
            tree = parser.parse(source.encode())
        except Exception:  # noqa: BLE001
            logger.warning("tree-sitter parse failed for %s", file_path, exc_info=True)
            return []

        symbols: list[Symbol] = []
        self._walk(tree.root_node, file_path, symbols, parent=None)
        return symbols

    # ------------------------------------------------------------------
    # Recursive walk
    # ------------------------------------------------------------------

    def _walk(
        self,
        node: Node,
        file_path: str,
        out: list[Symbol],
        parent: str | None,
    ) -> None:
        for child in node.children:
            ntype = child.type

            # Unwrap export wrappers transparently.
            if ntype == "export_statement":
                self._walk(child, file_path, out, parent)
                continue

            if ntype in ("function_declaration", "generator_function_declaration"):
                name = _field_text(child, "name")
                if name:
                    kind = self._func_kind(child, parent)
                    out.append(_sym(name, kind, child, file_path, parent))
                continue

            if ntype == "class_declaration":
                name = _field_text(child, "name")
                if name:
                    out.append(_sym(name, SymbolKind.CLASS, child, file_path, parent))
                    self._walk(child, file_path, out, parent=name)
                continue

            if ntype == "method_definition":
                name = _field_text(child, "name")
                if name:
                    kind = self._func_kind(child, parent, default=SymbolKind.METHOD)
                    out.append(_sym(name, kind, child, file_path, parent))
                continue

            if ntype == "import_statement":
                source = _import_source(child)
                if source:
                    out.append(_sym(source, SymbolKind.IMPORT, child, file_path, parent))
                continue

            if ntype in ("lexical_declaration", "variable_declaration"):
                self._handle_var_declaration(child, file_path, out, parent)
                continue

            # Recurse into everything else (if-blocks, etc.) to catch
            # nested declarations in module-init code.
            self._walk(child, file_path, out, parent)

    # ------------------------------------------------------------------
    # Variable / const declarations
    # ------------------------------------------------------------------

    def _handle_var_declaration(
        self,
        node: Node,
        file_path: str,
        out: list[Symbol],
        parent: str | None,
    ) -> None:
        is_const = node.type == "lexical_declaration" and _has_const_keyword(node)

        for decl in node.children:
            if decl.type != "variable_declarator":
                continue

            name_node = decl.child_by_field_name("name")
            value_node = decl.child_by_field_name("value")
            if name_node is None:
                continue

            name = name_node.text.decode() if name_node.text else None
            if not name:
                continue

            # Arrow / function expression → promote to function symbol.
            if value_node and value_node.type in ("arrow_function", "function_expression"):
                is_async = any(c.type == "async" for c in value_node.children)
                kind = SymbolKind.ASYNC_FUNCTION if is_async else SymbolKind.FUNCTION
                out.append(_sym(name, kind, node, file_path, parent))
            else:
                kind = SymbolKind.CONSTANT if is_const and name.isupper() else SymbolKind.VARIABLE
                out.append(_sym(name, kind, node, file_path, parent))

    # ------------------------------------------------------------------
    # Helpers
    # ------------------------------------------------------------------

    @staticmethod
    def _func_kind(
        node: Node,
        parent: str | None,
        default: SymbolKind = SymbolKind.FUNCTION,
    ) -> SymbolKind:
        is_async = any(c.type == "async" for c in node.children)
        if parent and default == SymbolKind.METHOD:
            return SymbolKind.METHOD
        return SymbolKind.ASYNC_FUNCTION if is_async else default


# ---------------------------------------------------------------------------
# Module-level helpers (not methods — they don't need ``self``).
# ---------------------------------------------------------------------------


def _ext_of(path: str) -> str:
    dot = path.rfind(".")
    return path[dot:] if dot != -1 else ""


def _field_text(node: Node, field: str) -> str | None:
    child = node.child_by_field_name(field)
    if child and child.text:
        return child.text.decode()
    return None


def _import_source(node: Node) -> str | None:
    src = node.child_by_field_name("source")
    if src and src.text:
        return src.text.decode().strip("'\"")
    return None


def _has_const_keyword(node: Node) -> bool:
    for child in node.children:
        if child.type == "const":
            return True
    return False


def _sym(
    name: str,
    kind: SymbolKind,
    node: Node,
    file_path: str,
    parent: str | None,
) -> Symbol:
    return Symbol(
        name=name,
        kind=kind,
        start_line=node.start_point[0] + 1,
        end_line=node.end_point[0] + 1,
        file_path=file_path,
        parent=parent,
    )
