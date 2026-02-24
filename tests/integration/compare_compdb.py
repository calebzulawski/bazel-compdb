#!/usr/bin/env python3
"""Validate compile_commands.json shape and key values."""

import json
import os
from pathlib import Path


def _validate(data: object) -> list[str]:
    entry = data[0]
    errors: list[str] = []

    directory = entry["directory"]
    if not isinstance(directory, str) or not directory:
        errors.append("missing or invalid 'directory'")

    file_value = str(entry["file"])
    output_value = str(entry["output"])
    args = entry["arguments"]

    has_compile_unix = "-c" in args and file_value in args
    has_compile_windows = "/c" in args and file_value in args
    has_unix_style = "-o" in args and file_value in args
    has_windows_style = any(a == f"/Fo{output_value}" for a in args)
    if not (has_compile_unix or has_compile_windows):
        errors.append(f"expected compile args (-c or /c) with source path {file_value}")
    if not (has_unix_style or has_windows_style):
        errors.append(
            f"expected either unix args (-o and {file_value}) or windows arg (/Fo{output_value})"
        )

    return errors


def main() -> int:
    try:
        workspace = os.environ["BUILD_WORKSPACE_DIRECTORY"]
        input_path = Path(workspace) / "compile_commands.json"
        data = json.loads(input_path.read_text(encoding="utf-8"))
    except Exception as exc:
        print(str(exc))
        return 1

    errors = _validate(data)
    if errors:
        print("compile_commands.json validation failed:")
        print("\n".join(f"- {e}" for e in errors))
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
