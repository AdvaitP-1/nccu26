"""Tests for the overlap detection service — the core analysis engine."""

from __future__ import annotations

import pytest

from app.core.overlap_service import detect_overlaps
from app.schemas import OverlapSeverity
from tests.conftest import (
    PY_CLASS_METHODS,
    PY_CLASS_METHODS_ALT,
    PY_FUNC_A,
    PY_FUNC_B,
    PY_FUNC_C,
    PY_IMPORTS_ONLY,
    PY_INVALID,
    PY_SEPARATE,
    PY_TOPLEVEL_VAR,
    TS_FUNC_A,
    TS_FUNC_B,
    make_changeset,
    make_request,
)


# -----------------------------------------------------------------------
# 1.  Two agents, same file, same function → critical
# -----------------------------------------------------------------------
class TestTwoAgentsSameSymbol:
    def test_overlapping_function_is_critical(self):
        req = make_request(
            make_changeset("alpha", [("src/auth.py", "python", PY_FUNC_A)]),
            make_changeset("beta", [("src/auth.py", "python", PY_FUNC_B)]),
        )
        overlaps = detect_overlaps(req)
        vt = [o for o in overlaps if o.symbol_name == "validate_token"]
        assert len(vt) >= 1
        assert any(o.severity == OverlapSeverity.CRITICAL for o in vt)


# -----------------------------------------------------------------------
# 2.  Two agents, same file, different functions → no dangerous overlap
# -----------------------------------------------------------------------
class TestTwoAgentsDifferentSymbols:
    def test_no_overlap_on_different_functions(self):
        req = make_request(
            make_changeset("alpha", [("src/utils.py", "python", PY_FUNC_A)]),
            make_changeset("beta", [("src/utils.py", "python", PY_SEPARATE)]),
        )
        overlaps = detect_overlaps(req)
        assert len(overlaps) == 0


# -----------------------------------------------------------------------
# 3.  Two agents, different files → no overlap
# -----------------------------------------------------------------------
class TestTwoAgentsDifferentFiles:
    def test_different_files_no_overlap(self):
        req = make_request(
            make_changeset("alpha", [("src/a.py", "python", PY_FUNC_A)]),
            make_changeset("beta", [("src/b.py", "python", PY_FUNC_B)]),
        )
        overlaps = detect_overlaps(req)
        assert len(overlaps) == 0


# -----------------------------------------------------------------------
# 4.  Three agents, selective pairwise overlaps
# -----------------------------------------------------------------------
class TestThreeAgentsPairwise:
    def test_three_agents_partial_overlap(self):
        req = make_request(
            make_changeset("alpha", [("src/auth.py", "python", PY_FUNC_A)]),
            make_changeset("beta", [("src/auth.py", "python", PY_FUNC_B)]),
            make_changeset("gamma", [("src/auth.py", "python", PY_FUNC_C)]),
        )
        overlaps = detect_overlaps(req)
        pairs = {(o.agent_a, o.agent_b) for o in overlaps if o.symbol_name == "validate_token"}
        # All three pairs should appear because all three define validate_token
        assert ("alpha", "beta") in pairs
        assert ("alpha", "gamma") in pairs
        assert ("beta", "gamma") in pairs

    def test_three_agents_only_two_overlap(self):
        req = make_request(
            make_changeset("alpha", [("src/auth.py", "python", PY_FUNC_A)]),
            make_changeset("beta", [("src/auth.py", "python", PY_FUNC_B)]),
            make_changeset("gamma", [("src/other.py", "python", PY_SEPARATE)]),
        )
        overlaps = detect_overlaps(req)
        pairs = {(o.agent_a, o.agent_b) for o in overlaps}
        assert ("alpha", "beta") in pairs
        assert ("alpha", "gamma") not in pairs
        assert ("beta", "gamma") not in pairs


# -----------------------------------------------------------------------
# 5.  Four agents on one file → correct pairwise count
# -----------------------------------------------------------------------
class TestFourAgentsOneFile:
    def test_four_agents_all_same_function(self):
        req = make_request(
            make_changeset("a1", [("f.py", "python", PY_FUNC_A)]),
            make_changeset("a2", [("f.py", "python", PY_FUNC_B)]),
            make_changeset("a3", [("f.py", "python", PY_FUNC_C)]),
            make_changeset("a4", [("f.py", "python", "def validate_token(x):\n    return x\n")]),
        )
        overlaps = detect_overlaps(req)
        vt = [o for o in overlaps if o.symbol_name == "validate_token"]
        pairs = {(o.agent_a, o.agent_b) for o in vt}
        # C(4,2) = 6 unique pairs
        assert len(pairs) == 6


# -----------------------------------------------------------------------
# 6.  Same class, different methods → only matching methods overlap
# -----------------------------------------------------------------------
class TestClassMethodOverlap:
    def test_only_shared_method_overlaps(self):
        req = make_request(
            make_changeset("alpha", [("svc.py", "python", PY_CLASS_METHODS)]),
            make_changeset("beta", [("svc.py", "python", PY_CLASS_METHODS_ALT)]),
        )
        overlaps = detect_overlaps(req)
        names = {o.symbol_name for o in overlaps}
        assert "start" in names  # both define start
        assert "Service" in names  # both define the class
        assert "stop" not in names  # only alpha has stop
        assert "health" not in names  # only beta has health


# -----------------------------------------------------------------------
# 7.  Import-only overlap
# -----------------------------------------------------------------------
class TestImportOverlap:
    def test_shared_import_is_not_critical_unless_overlapping(self):
        req = make_request(
            make_changeset("alpha", [("m.py", "python", PY_IMPORTS_ONLY)]),
            make_changeset("beta", [("m.py", "python", PY_IMPORTS_ONLY)]),
        )
        overlaps = detect_overlaps(req)
        # Imports at same lines → critical (they are overlapping)
        assert len(overlaps) > 0
        for o in overlaps:
            assert o.symbol_kind == "import"


# -----------------------------------------------------------------------
# 8.  Top-level variable overlap
# -----------------------------------------------------------------------
class TestVariableOverlap:
    def test_variable_overlap(self):
        req = make_request(
            make_changeset("alpha", [("cfg.py", "python", PY_TOPLEVEL_VAR)]),
            make_changeset("beta", [("cfg.py", "python", PY_TOPLEVEL_VAR)]),
        )
        overlaps = detect_overlaps(req)
        names = {o.symbol_name for o in overlaps}
        assert "MAX_RETRIES" in names
        assert "timeout" in names


# -----------------------------------------------------------------------
# 9.  Invalid Python content — graceful degradation
# -----------------------------------------------------------------------
class TestInvalidPython:
    def test_invalid_python_does_not_crash(self):
        req = make_request(
            make_changeset("alpha", [("bad.py", "python", PY_INVALID)]),
            make_changeset("beta", [("bad.py", "python", PY_FUNC_A)]),
        )
        overlaps = detect_overlaps(req)
        # alpha's file fails to parse, so no shared symbols → no overlaps
        assert isinstance(overlaps, list)

    def test_valid_files_still_process_alongside_invalid(self):
        req = make_request(
            make_changeset("alpha", [
                ("bad.py", "python", PY_INVALID),
                ("good.py", "python", PY_FUNC_A),
            ]),
            make_changeset("beta", [
                ("good.py", "python", PY_FUNC_B),
            ]),
        )
        overlaps = detect_overlaps(req)
        assert any(o.file_path == "good.py" for o in overlaps)


# -----------------------------------------------------------------------
# 10.  Invalid TS content — graceful degradation
# -----------------------------------------------------------------------
class TestInvalidTS:
    def test_invalid_ts_does_not_crash(self):
        req = make_request(
            make_changeset("alpha", [("bad.ts", "typescript", "export function {{{")]),
            make_changeset("beta", [("bad.ts", "typescript", TS_FUNC_A)]),
        )
        # Should not raise — tree-sitter produces partial trees
        overlaps = detect_overlaps(req)
        assert isinstance(overlaps, list)


# -----------------------------------------------------------------------
# 11.  Empty changesets
# -----------------------------------------------------------------------
class TestEmptyInputs:
    def test_empty_changesets_list(self):
        req = make_request()
        overlaps = detect_overlaps(req)
        assert overlaps == []

    def test_changeset_with_empty_files(self):
        req = make_request(
            make_changeset("alpha", []),
            make_changeset("beta", []),
        )
        overlaps = detect_overlaps(req)
        assert overlaps == []

    def test_single_changeset_no_pair(self):
        req = make_request(
            make_changeset("alpha", [("f.py", "python", PY_FUNC_A)]),
        )
        overlaps = detect_overlaps(req)
        assert overlaps == []


# -----------------------------------------------------------------------
# 12.  Unsupported file extension
# -----------------------------------------------------------------------
class TestUnsupportedExtension:
    def test_rust_file_ignored_gracefully(self):
        req = make_request(
            make_changeset("alpha", [("lib.rs", "rust", "fn main() {}")]),
            make_changeset("beta", [("lib.rs", "rust", "fn main() {}")]),
        )
        overlaps = detect_overlaps(req)
        assert overlaps == []


# -----------------------------------------------------------------------
# 13.  Mixed language batch — Python + TS in one request
# -----------------------------------------------------------------------
class TestMixedLanguages:
    def test_python_and_ts_in_one_request(self):
        req = make_request(
            make_changeset("alpha", [
                ("auth.py", "python", PY_FUNC_A),
                ("api.ts", "typescript", TS_FUNC_A),
            ]),
            make_changeset("beta", [
                ("auth.py", "python", PY_FUNC_B),
                ("api.ts", "typescript", TS_FUNC_B),
            ]),
        )
        overlaps = detect_overlaps(req)
        files = {o.file_path for o in overlaps}
        assert "auth.py" in files
        assert "api.ts" in files


# -----------------------------------------------------------------------
# 14.  Many agents, severity distinctions
# -----------------------------------------------------------------------
class TestSeverityDistinctions:
    def test_separated_symbols_are_medium(self):
        # Two agents both define validate_token but at very different line ranges
        src_a = "\n" * 50 + "def validate_token(t):\n    return t\n"
        src_b = "def validate_token(t):\n    return t\n"
        req = make_request(
            make_changeset("alpha", [("f.py", "python", src_a)]),
            make_changeset("beta", [("f.py", "python", src_b)]),
        )
        overlaps = detect_overlaps(req)
        vt = [o for o in overlaps if o.symbol_name == "validate_token"]
        assert len(vt) == 1
        assert vt[0].severity in (OverlapSeverity.MEDIUM, OverlapSeverity.HIGH)
