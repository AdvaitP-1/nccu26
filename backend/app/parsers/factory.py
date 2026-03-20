"""Parser factory — resolves a file extension to the correct parser."""

from __future__ import annotations

import logging
from pathlib import PurePosixPath

from app.parsers.base import BaseParser
from app.parsers.python_parser import PythonParser
from app.parsers.ts_parser import TypeScriptParser

logger = logging.getLogger(__name__)

# Singletons — parsers are stateless so one instance per language is enough.
_PYTHON = PythonParser()
_TYPESCRIPT = TypeScriptParser()

_EXTENSION_MAP: dict[str, BaseParser] = {}
for _parser in (_PYTHON, _TYPESCRIPT):
    for _ext in _parser.supported_extensions():
        _EXTENSION_MAP[_ext] = _parser


def get_parser(file_path: str) -> BaseParser | None:
    """Return the appropriate parser for *file_path*, or ``None`` if unsupported."""
    ext = PurePosixPath(file_path).suffix.lower()
    return _EXTENSION_MAP.get(ext)
