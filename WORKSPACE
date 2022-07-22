load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "16e9fca53ed6bd4ff4ad76facc9b7b651a89db1689a2877d6fd7b82aa824e366",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.34.0/rules_go-v0.34.0.zip",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.34.0/rules_go-v0.34.0.zip",
    ],
)

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")

go_rules_dependencies()

go_register_toolchains(version = "1.18.4")

http_archive(
    name = "bazel_gazelle",
    sha256 = "501deb3d5695ab658e82f6f6f549ba681ea3ca2a5fb7911154b5aa45596183fa",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.26.0/bazel-gazelle-v0.26.0.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.26.0/bazel-gazelle-v0.26.0.tar.gz",
    ],
)

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")
load("//:go_repos.bzl", "go_repositories")

# gazelle:repository_macro go_repos.bzl%go_repositories
go_repositories()

gazelle_dependencies()

http_archive(
    name = "com_google_protobuf",
    sha256 = "1e958b62debbb46ebefa16d848741d17c88dc018dd57b729c0cd58965380f3f8",
    strip_prefix = "protobuf-21.3",
    urls = [
        "https://github.com/protocolbuffers/protobuf/archive/v21.3.tar.gz",
    ],
)

load("@com_google_protobuf//:protobuf_deps.bzl", "protobuf_deps")

protobuf_deps()

BAZEL_ZIG_CC_VERSION = "v0.8.2"

http_archive(
    name = "bazel-zig-cc",
    sha256 = "a91a1769d98d6ec0166874a83e9d8023a5a38a0bd02208035b281ff38adf7029",
    strip_prefix = "bazel-zig-cc-c122a0d1c70503b8ea4f95f80ff21ccbc7fcdd42",
    urls = ["https://github.com/gmirror/bazel-zig-cc/archive/c122a0d1c70503b8ea4f95f80ff21ccbc7fcdd42.tar.gz".format(BAZEL_ZIG_CC_VERSION)],
)

load("@bazel-zig-cc//toolchain:defs.bzl", zig_toolchains = "toolchains")

zig_toolchains()

register_toolchains(
    "@zig_sdk//toolchain:aarch64-linux-gnu.2.28",
    "@zig_sdk//toolchain:x86_64-linux-gnu.2.28",
)

http_archive(
    name = "io_bazel_rules_docker",
    sha256 = "b1e80761a8a8243d03ebca8845e9cc1ba6c82ce7c5179ce2b295cd36f7e394bf",
    urls = ["https://github.com/bazelbuild/rules_docker/releases/download/v0.25.0/rules_docker-v0.25.0.tar.gz"],
)

load("@io_bazel_rules_docker//toolchains/docker:toolchain.bzl", docker_toolchain_configure = "toolchain_configure")

docker_toolchain_configure(
    name = "docker_config",
    # client_config = "//tools/docker:config.json",
    # cred_helpers = ["//tools/docker:docker-credential-cloudflared.sh"],
    # The toolchain may complain that xz is missing, but we don't use it.
    xz_path = "/usr/bin/false",
)

load("@io_bazel_rules_docker//repositories:repositories.bzl", container_repositories = "repositories")

container_repositories()

load("@io_bazel_rules_docker//repositories:deps.bzl", container_deps = "deps")

container_deps()

load("@io_bazel_rules_docker//go:image.bzl", _go_image_repos = "repositories")

_go_image_repos()

http_archive(
    name = "com_github_bazelbuild_buildtools",
    sha256 = "ae34c344514e08c23e90da0e2d6cb700fcd28e80c02e23e4d5715dddcb42f7b3",
    strip_prefix = "buildtools-4.2.2",
    urls = [
        "https://github.com/bazelbuild/buildtools/archive/refs/tags/4.2.2.tar.gz",
    ],
)
