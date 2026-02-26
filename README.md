# bazel-compdb

Generate `compile_commands.json` from Bazel.

## Install

You may either download a prebuilt binary or add `bazel-compdb` to your `MODULE.bazel`.

### Download a release
Download a prebuilt binary from the [latest release](https://github.com/calebzulawski/bazel-compdb/releases/latest).

### Use via Bazel
Add `bazel-compdb` to your `MODULE.bazel`:

```starlark
bazel_dep(name = "bazel-compdb", version = "<version>")
```

Run it with:

```bash
bazel run @bazel-compdb
```

## Usage

`bazel-compdb` accepts arguments in either of these forms:

```bash
bazel-compdb [arguments] -- [bazel arguments]
bazel-compdb [bazel arguments]
```

When running through Bazel:

```bash
bazel run @bazel-compdb -- [arguments] -- [bazel arguments]
```

Examples:

```bash
# Default target set (//...)
bazel run @bazel-compdb

# Pass Bazel build args through unchanged
bazel run @bazel-compdb -- -c dbg //foo:bar

# Write output to a custom location
bazel run @bazel-compdb -- -o out/compile_commands.json -- //...
```
