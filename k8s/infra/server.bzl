load("@rules_go//go:def.bzl", "go_cross_binary", _go_binary = "go_binary")
load("//k8s/infra:image.bzl", "image")

def go_binary(name, embed, visibility, **kwargs):
    """
    Wrapper macro

    Args:
        name: Name of the image and related targets.
        embed: List of source files (typically the built binary) to include in the image.
        visibility: Base image to use for the OCI image.
        **kwargs: All other arguments passed to go_binary

    This macro:
      1. Builds golang binary
      2. Builds cross compiled golang binary
      3. Packages the cross compiled binary into a tar layer.
      4. Builds an OCI image using the specified base and entrypoint.
      5. Generates a unique, immutable tag based on the image's sha256 digest.
      6. Pushes the image to a local registry with both the sha256 tag and the 'latest' tag.
      7. Provides a multirun target to push both tags in one command.
    """
    _go_binary(
        name = name,
        embed = embed,
        visibility = visibility,
        **kwargs
    )

    cross_name = "%s_cross" % name
    go_cross_binary(
        name = cross_name,
        compilation_mode = "opt",
        platform = "@rules_go//go/toolchain:linux_amd64",
        target = ":%s" % name,
    )

    image(
        name = name,
        srcs = [":%s" % cross_name],
        base = "@distroless_base",
        entrypoint = ["/%s" % cross_name],
        exposed_ports = [
            "25000",
            "26000",
        ],
    )
