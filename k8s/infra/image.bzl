# auto add latest tag
load("@rules_multirun//:defs.bzl", "command", "multirun")

# OCI Image Rules
load("@rules_oci//oci:defs.bzl", "oci_image", "oci_push")

# OCI Container Rules
load("@rules_pkg//pkg:tar.bzl", "pkg_tar")

def _build_sha265_tag_impl(ctx):
    # Both the input and output files are specified by the BUILD file.
    in_file = ctx.file.input
    out_file = ctx.outputs.output

    # No need to return anything telling Bazel to build `out_file` when
    # building this target -- It's implied because the output is declared
    # as an attribute rather than with `declare_file()`.
    ctx.actions.run_shell(
        inputs = [in_file],
        outputs = [out_file],
        arguments = [in_file.path, out_file.path],
        command = "sed -n 's/.*sha256:\\([[:alnum:]]\\{7\\}\\).*/\\1/p' < \"$1\" > \"$2\"",
    )

build_sha265_tag = rule(
    doc = "Extracts a 7 characters long short hash from the image digest.",
    implementation = _build_sha265_tag_impl,
    attrs = {
        "image": attr.label(
            allow_single_file = True,
            mandatory = True,
        ),
        "input": attr.label(
            allow_single_file = True,
            mandatory = True,
            doc = "The image digest file. Usually called image.json.sha256",
        ),
        "output": attr.output(
            doc = "The generated tag file. Usually named _tag.txt",
        ),
    },
)

def image(name, srcs, base, entrypoint, exposed_ports = []):
    """
    Builds and publishes an OCI image for a given binary.

    Args:
        name: Name of the image and related targets.
        srcs: List of source files (typically the built binary) to include in the image.
        base: Base image to use for the OCI image.
        entrypoint: List specifying the entrypoint command for the container.
        exposed_ports: Optional list of ports to expose from the container.

    This macro:
      1. Packages the binary into a tar layer.
      2. Builds an OCI image using the specified base and entrypoint.
      3. Generates a unique, immutable tag based on the image's sha256 digest.
      4. Pushes the image to a local registry with both the sha256 tag and the 'latest' tag.
      5. Provides a multirun target to push both tags in one command.
    """

    # 1) Compress the Rust binary to tar
    pkg_tar(
        name = "{}_tar".format(name),
        srcs = srcs,
    )

    # 2) Build container image
    # https://github.com/bazel-contrib/rules_oci/blob/main/docs/image.md
    oci_image(
        name = "{}_image".format(name),
        base = base,
        entrypoint = entrypoint,
        exposed_ports = exposed_ports,
        tars = ["{}_tar".format(name)],
        visibility = ["//visibility:public"],
    )

    # 3) Build an unique and immutable image tag
    build_sha265_tag(
        name = "{}_remote_tag".format(name),
        image = "{}_image".format(name),
        input = "{}_image.json.sha256".format(name),
        output = "_tag.txt",
    )

    # 4) Define a registry to publish the image
    # https://github.com/bazel-contrib/rules_oci/blob/main/docs/push.md)
    oci_push(
        name = "{}_push_sha256_tag".format(name),
        image = "{}_image".format(name),
        remote_tags = "{}_remote_tag".format(name),
        repository = "localhost:5001/{}".format(name),
        visibility = ["//visibility:public"],
    )

    # 4) Define a registry to publish the image with latest tag
    # https://github.com/bazel-contrib/rules_oci/blob/main/docs/push.md)
    oci_push(
        name = "{}_push_latest_tag".format(name),
        image = "{}_image".format(name),
        remote_tags = ["latest"],
        repository = "localhost:5001/{}".format(name),
        visibility = ["//visibility:public"],
    )

    multirun(
        name = "{}_push".format(name),
        commands = ["{}_push_sha256_tag".format(name), "{}_push_latest_tag".format(name)],
    )
