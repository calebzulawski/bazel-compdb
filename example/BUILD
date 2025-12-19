cc_library(
    name = "greeter",
    srcs = ["src/greeter.cc"],
    hdrs = ["src/greeter.h"],
    visibility = ["//visibility:public"],
)

cc_binary(
    name = "hello_app",
    srcs = ["src/main.cc"],
    deps = [":greeter"],
)
