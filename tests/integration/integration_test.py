#!/usr/bin/env python3
"""Runs the integration test suite without nesting Bazel invocations."""

from __future__ import annotations

import argparse
import difflib
import json
import os
import platform
import re
import subprocess
import sys
from pathlib import Path
from typing import Optional


def _golden_path(workspace: Path) -> Path:
    system = platform.system().lower()
    return workspace / "tests" / "integration" / f"compile_commands.{system}.json"

_DROP_FLAG_PREFIXES = ("-W", "-f", "/w", "/z")


def _sanitize_string(value: str, tmp_dir: str) -> str:
    sanitized = value.replace(tmp_dir, "TEST_WORKSPACE")
    sanitized = re.sub(r"bazel-out/[A-Za-z0-9._-]+/bin", "bazel-out/TARGET/bin", sanitized)
    return sanitized


def _should_drop_flag(value: str) -> bool:
    return value.startswith(_DROP_FLAG_PREFIXES)


def _sanitize_entry(entry, tmp_dir: str):
    if isinstance(entry, str):
        return _sanitize_string(entry, tmp_dir)
    if isinstance(entry, list):
        sanitized_items = []
        saw_compiler = False
        for item in entry:
            sanitized_item = _sanitize_entry(item, tmp_dir)
            if isinstance(sanitized_item, str):
                if _should_drop_flag(sanitized_item):
                    continue
                if not saw_compiler:
                    sanitized_item = "TOOLCHAIN_COMPILER"
                    saw_compiler = True
            sanitized_items.append(sanitized_item)
        return sanitized_items
    if isinstance(entry, dict):
        return {key: _sanitize_entry(value, tmp_dir) for key, value in entry.items()}
    return entry


def _workspace_root() -> Path:
    return Path(os.environ["BUILD_WORKSPACE_DIRECTORY"])


def _bazel_compdb_path() -> Path:
    manifest = os.environ["RUNFILES_MANIFEST_FILE"]
    with Path(manifest).open("r", encoding="utf-8") as handle:
        for line in handle:
            line = line.rstrip()
            if not line or " " not in line:
                continue
            key, value = line.split(" ", 1)
            if key.endswith(("bazel-compdb_/bazel-compdb.exe", "bazel-compdb_/bazel-compdb", "bazel-compdb.exe", "bazel-compdb")):
                path = Path(value)
                if path.exists():
                    return path
    raise FileNotFoundError("bazel-compdb binary not found in runfiles manifest.")


def _generate_compile_commands(workspace: Path) -> str:
    bazel_compdb = _bazel_compdb_path()
    example_project = workspace / "tests" / "integration" / "example_project"
    compile_commands = example_project / "compile_commands.json"
    compile_commands.unlink(missing_ok=True)
    subprocess.run(
        [str(bazel_compdb), "--", "//:hello_world"],
        cwd=example_project,
        check=True,
    )
    if not compile_commands.is_file():
        raise FileNotFoundError(f"{compile_commands} was not produced by bazel-compdb")
    with compile_commands.open("r", encoding="utf-8") as handle:
        data = json.load(handle)
    sanitized = _sanitize_entry(data, example_project.resolve().as_posix())
    compile_commands.unlink(missing_ok=True)
    return json.dumps(sanitized, indent=2) + "\n"


def _write_if_requested(content: str, destination: Optional[Path]) -> None:
    if destination is None:
        return
    destination.parent.mkdir(parents=True, exist_ok=True)
    destination.write_text(content, encoding="utf-8")


def main() -> None:
    parser = argparse.ArgumentParser(description="Run integration tests via bazel-compdb.")
    parser.add_argument(
        "-o",
        "--output",
        help="Optional path (relative to the workspace) to write the sanitized compile_commands.json.",
    )
    args = parser.parse_args()

    workspace = _workspace_root()
    sanitized_content = _generate_compile_commands(workspace)

    output_path = (workspace / args.output).resolve() if args.output else None
    _write_if_requested(sanitized_content, output_path)

    golden_path = _golden_path(workspace)
    if not golden_path.exists():
        raise FileNotFoundError(f"Golden file {golden_path} does not exist. Supply -o to create it.")

    expected = golden_path.read_text(encoding="utf-8")
    if sanitized_content == expected:
        print(f"Integration compile commands match {golden_path}")
        return

    raise RuntimeError(
        "Integration compile commands differ from the golden file.\n"
        + "".join(difflib.unified_diff(
            expected.splitlines(keepends=True),
            sanitized_content.splitlines(keepends=True),
            fromfile=str(golden_path),
            tofile="sanitized_compile_commands.json",
        ))
    )


if __name__ == "__main__":
    main()
