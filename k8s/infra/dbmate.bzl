"""Bazel rules for building and pushing dbmate migration images"""

load("@rules_multirun//:defs.bzl", "multirun")
load("@rules_oci//oci:defs.bzl", "oci_image", "oci_image_index", "oci_push")
load("@rules_pkg//pkg:tar.bzl", "pkg_tar")
load("//k8s/infra:image.bzl", "build_sha265_tag")

def dbmate_image(
        name,
        migrations,
        package_dir = "/migrations",
        repositories = ["registry.localhost"],
        visibility = None):
    """Creates a dbmate image with migrations bundled.

    This macro creates all necessary targets for building and pushing a dbmate
    migration image:
    - pkg_tar for packaging migrations
    - multi-platform OCI images (linux/amd64, linux/arm64)
    - oci_image_index for multi-platform support
    - sha256 tag generation
    - push targets for each repository
    - multirun target to push all tags

    Args:
        name: Base name for the targets (e.g., "dbmate")
        migrations: List of migration files or filegroup label
        package_dir: Directory in the image where migrations will be stored (default: "/migrations")
        repositories: List of container registries to push to (default: ["registry.localhost"])
        visibility: Visibility for the main targets

    Generated targets:
        {name}_image: multi-platform OCI image
        {name}_remote_tag: sha256 tag file
        {name}_push: multirun target to push to all repositories
    """

    # Package migration files into a tar
    pkg_tar(
        name = "{}_migrations_tar".format(name),
        srcs = migrations,
        package_dir = package_dir,
    )

    # Build OCI image with dbmate + migrations for each platform
    oci_image(
        name = "{}_linux_amd64".format(name),
        base = "@dbmate_linux_amd64",
        tars = [":{}_migrations_tar".format(name)],
        workdir = "/",
    )

    oci_image(
        name = "{}_linux_arm64".format(name),
        base = "@dbmate_linux_arm64",
        tars = [":{}_migrations_tar".format(name)],
        workdir = "/",
    )

    # Create multi-platform image index
    oci_image_index(
        name = "{}_image".format(name),
        images = [
            ":{}_linux_amd64".format(name),
            ":{}_linux_arm64".format(name),
        ],
    )

    # Generate sha256 tag
    build_sha265_tag(
        name = "{}_remote_tag".format(name),
        image = ":{}_image".format(name),
        input = ":{}_image.json.sha256".format(name),
        output = "{}_tag.txt".format(name),
    )

    # Create push targets for each repository
    push_targets = []
    for repo in repositories:
        # Normalize repository name for target naming (replace special chars with _)
        repo_safe = repo.replace(":", "_").replace("/", "_").replace(".", "_")

        # Push with sha256 tag
        sha_target = "{}_{}_push_sha256_tag".format(name, repo_safe)
        oci_push(
            name = sha_target,
            image = ":{}_image".format(name),
            remote_tags = ":{}_remote_tag".format(name),
            repository = "{}/{}".format(repo, name),
        )
        push_targets.append(":{}".format(sha_target))

        # Push with latest tag
        latest_target = "{}_{}_push_latest_tag".format(name, repo_safe)
        oci_push(
            name = latest_target,
            image = ":{}_image".format(name),
            remote_tags = ["latest"],
            repository = "{}/{}".format(repo, name),
        )
        push_targets.append(":{}".format(latest_target))

    # Multi-run target to push all tags to all repositories
    multirun(
        name = "{}_push".format(name),
        commands = push_targets,
        visibility = visibility or ["//visibility:public"],
    )
