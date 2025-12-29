#!/usr/bin/env python3
"""Normalize compile_commands.json for snapshot comparison."""

from __future__ import annotations

import argparse
import json
import re
from pathlib import Path


def _normalize_string(value: str) -> str:
    normalized = value.replace("\\", "/")
    normalized = re.sub(
        r"(?:[A-Za-z]:)?[^\n]*?/tests/integration/example_project",
        "TEST_WORKSPACE",
        normalized,
    )
    normalized = re.sub(r"bazel-out/[A-Za-z0-9._-]+/bin", "bazel-out/TARGET/bin", normalized)
    return normalized


def _normalize_entry(entry):
    if isinstance(entry, dict):
        return _normalize_compile_command(entry)
    if isinstance(entry, list):
        return [_normalize_entry(item) for item in entry]
    if isinstance(entry, str):
        return _normalize_string(entry)
    return entry


def _normalize_file(value: str) -> str:
    normalized = _normalize_string(value)
    if not normalized:
        return "SOURCE_FILE"
    return normalized.split("/")[-1]


def _normalize_output(value: str) -> str:
    _normalize_string(value)
    return "OBJECT_FILE"


def _normalize_compile_command(entry: dict) -> dict:
    directory = "TEST_WORKSPACE" if "directory" in entry else None
    file_value = _normalize_file(entry.get("file", "")) if "file" in entry else None
    output_value = _normalize_output(entry.get("output", "")) if "output" in entry else None
    arguments = None
    if "arguments" in entry or "command" in entry:
        arguments = [
            "TOOLCHAIN_COMPILER",
            "-c",
            file_value or "SOURCE_FILE",
            "-o",
            output_value or "OBJECT_FILE",
        ]

    normalized = {}
    if directory is not None:
        normalized["directory"] = directory
    if arguments is not None:
        normalized["arguments"] = arguments
    if file_value is not None:
        normalized["file"] = file_value
    if output_value is not None:
        normalized["output"] = output_value

    return normalized


def main() -> int:
    parser = argparse.ArgumentParser(description="Normalize compile_commands.json.")
    parser.add_argument("input", help="Path to the raw compile_commands.json.")
    parser.add_argument("output", help="Path to write the normalized output.")
    args = parser.parse_args()

    with Path(args.input).open("r", encoding="utf-8") as handle:
        data = json.load(handle)

    normalized = _normalize_entry(data)
    content = json.dumps(normalized, indent=2) + "\n"
    Path(args.output).write_text(content, encoding="utf-8")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
