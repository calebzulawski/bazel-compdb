#!/usr/bin/env python3
import json
import re
import sys
from pathlib import Path


def sanitize_string(value: str, tmp_dir: str) -> str:
    sanitized = value.replace(tmp_dir, "TEST_WORKSPACE")
    sanitized = re.sub(r"bazel-out/[A-Za-z0-9._-]+/bin", "bazel-out/TARGET/bin", sanitized)
    return sanitized


def should_drop_flag(value: str) -> bool:
    return any(value.startswith(flag) for flag in ["-W", "-f", "/w", "/z"])


def sanitize_entry(entry, tmp_dir: str):
    if isinstance(entry, str):
        return sanitize_string(entry, tmp_dir)
    if isinstance(entry, list):
        sanitized_items = []
        saw_compiler = False
        for item in entry:
            sanitized_item = sanitize_entry(item, tmp_dir)
            if isinstance(sanitized_item, str):
                if should_drop_flag(sanitized_item):
                    continue
                if not saw_compiler:
                    sanitized_item = "TOOLCHAIN_COMPILER"
                    saw_compiler = True
            sanitized_items.append(sanitized_item)
        return sanitized_items
    if isinstance(entry, dict):
        return {key: sanitize_entry(value, tmp_dir) for key, value in entry.items()}
    return entry


def main() -> None:
    if len(sys.argv) != 4:
        print(
            "Usage: sanitize_compile_commands.py <tmp_dir> <input_json> <output_json>",
            file=sys.stderr,
        )
        sys.exit(1)

    tmp_dir = Path(sys.argv[1]).resolve().as_posix()
    input_path = Path(sys.argv[2])
    output_path = Path(sys.argv[3])

    with input_path.open("r", encoding="utf-8") as infile:
        data = json.load(infile)

    sanitized = sanitize_entry(data, tmp_dir)

    with output_path.open("w", encoding="utf-8") as outfile:
        json.dump(sanitized, outfile, indent=2)
        outfile.write("\n")


if __name__ == "__main__":
    main()
