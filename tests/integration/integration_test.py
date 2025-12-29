#!/usr/bin/env python3
"""Runs the integration test suite without nesting Bazel invocations."""

from __future__ import annotations

import os
import shutil
import subprocess
from pathlib import Path

from python.runfiles import runfiles


def _snapshot_output_path() -> Path:
    output_root = os.environ.get("SNAPSHOT_OUTPUTS_DIR")
    assert output_root
    return Path(output_root) / "compile_commands.json"


def _runfiles() -> runfiles:
    runfiles_ctx = runfiles.Create()
    assert runfiles_ctx
    return runfiles_ctx


def _rlocation(path: str) -> Path:
    resolved = _runfiles().Rlocation(path)
    assert resolved
    resolved_path = Path(resolved)
    assert resolved_path.exists()
    return resolved_path


def _example_project_path() -> Path:
    workspace = os.environ.get("TEST_WORKSPACE", "_main")
    for marker in ("BUILD.bazel", "MODULE.bazel"):
        marker_path = f"{workspace}/tests/integration/example_project/{marker}"
        try:
            return _rlocation(marker_path).parent
        except AssertionError:
            continue
    assert False


def _bazel_compdb_path() -> Path:
    workspace = os.environ.get("TEST_WORKSPACE", "_main")
    candidates = [
        f"{workspace}/bazel-compdb",
        f"{workspace}/bazel-compdb.exe",
        f"{workspace}/bazel-compdb_/bazel-compdb",
        f"{workspace}/bazel-compdb_/bazel-compdb.exe",
    ]
    for candidate in candidates:
        resolved = _runfiles().Rlocation(candidate)
        if not resolved:
            continue
        resolved_path = Path(resolved)
        if resolved_path.exists():
            return resolved_path
    assert False


def _run_bazel_compdb(example_project: Path) -> Path:
    bazel_compdb = _bazel_compdb_path()
    compile_commands = example_project / "compile_commands.json"
    compile_commands.unlink(missing_ok=True)
    subprocess.run(
        [str(bazel_compdb), "--", "//:hello_world"],
        cwd=example_project,
        check=True,
    )
    assert compile_commands.is_file()
    return compile_commands


def main() -> None:
    example_project = _example_project_path()
    compile_commands = _run_bazel_compdb(example_project)

    output_path = _snapshot_output_path()
    output_path.parent.mkdir(parents=True, exist_ok=True)
    shutil.copy2(compile_commands, output_path)


if __name__ == "__main__":
    main()
