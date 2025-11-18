"""
Defines the k8s targets.
"""

load("@bazel_skylib//rules:expand_template.bzl", "expand_template")

# load("@rules_k8s//k8s:objects.bzl", "k8s_objects")
load("@rules_kustomize//kustomize:kustomize.bzl", "kustomization", "kustomized_resources")
load("@rules_shell//shell:sh_binary.bzl", "sh_binary")

def create_substitutions(name, namespace):
    return {"{{name}}": name, "{{namespace}}": namespace}

def target(name, dir, images = [], enable_helm = False, extra_files = [], substitutions = {}, sync_group = "apps"):
    """
    Defines a target configuration for Kubernetes deployments.

    Args:
        name: The name of the target/application.
        dir: The directory containing the k8s manifests or configuration.
        images: Images to use in the k8s resource
        enable_helm: Boolean indicating if Helm is enabled for this target.
        extra_files: List of extra files to include in the target.
        substitutions: Extra dictionary of yaml substitutions
        sync_group: synchronization level

    Returns:
        A dictionary containing the target's configuration.
    """
    return {
        "name": name,
        "dir": dir,
        "images": images,
        "enable_helm": enable_helm,
        "extra_files": extra_files,
        "substitutions": substitutions,
        "sync_group": sync_group,
    }

def cluster(name, environment, path, url):
    """
    Defines a cluster configuration for Kubernetes deployments.

    Args:
        name: The name of the cluster.
        environment: The environment type (e.g., dev, staging, production).
        path: The path or directory associated with the cluster's configuration.
        url: The service URL or domain for the cluster.

    Returns:
        A dictionary containing the cluster's configuration.
    """
    return {
        "name": name,
        "environment": environment,
        "path": path,
        "url": url,
    }

CLUSTERS = [
    cluster(
        name = "dev",
        environment = "dev",
        path = "",
        url = "",
    ),
    cluster(
        name = "prod-eu-west-1",
        environment = "prod",
        path = "eu-west/1",
        url = "",
    ),
    cluster(
        name = "staging-eu-west-1",
        environment = "staging",
        path = "eu-west/1",
        url = "",
    ),
]

def deploy_targets(name, targets, dir_prefix = ""):
    """This loop generates yamls, argo application files and k8s bazel targets for each application and each relevant cluster.

    Args:
        name: Name
        targets: The target dictionaries to generate Bazel targets for.
        dir_prefix: Prefix of the directory to search in

    Returns:
        Returns the targets.
    """

    # print(targets)
    # print(CLUSTERS)
    for target in targets:
        for cluster in CLUSTERS:
            resource_file = get_k8s_resource_file(target["dir"], cluster["path"], cluster["environment"], dir_prefix)

            # print("resource file {}".format(resource_file))
            if resource_file == None:
                continue

            # Reflects the kustomization.yaml file
            kustomization(
                name = "{}-{}-kustomization".format(target["name"], cluster["name"]),
                srcs = native.glob(["{}/**/*.yaml".format(target["dir"])]) + target["extra_files"],
                file = resource_file,
                requires_helm = target["enable_helm"],
                visibility = ["//visibility:private"],
            )

            # Invocation of build kustomization.yaml command
            kustomized_resources(
                name = "{}-{}-kustomized-resource".format(target["name"], cluster["name"]),
                kustomization = "{}-{}-kustomization".format(target["name"], cluster["name"]),
                load_restrictor = "None",
                visibility = ["//visibility:private"],
            )

            substitutions = {
                "{{cluster_name}}": cluster["name"],
                "{{cluster_env}}": cluster["environment"],
                "{{cluster_url}}": cluster["url"],
                "{{cluster_path}}": cluster["path"],
            }
            substitutions.update(target["substitutions"])

            expanded_name = "%s-%s-expanded" % (target["name"], cluster["name"])
            expand_template(
                name = expanded_name,
                out = expanded_name,
                substitutions = substitutions,
                template = ":%s-%s-kustomized-resource" % (target["name"], cluster["name"]),
            )

            sh_binary(
                name = "%s-%s-print" % (target["name"], cluster["name"]),
                srcs = ["//k8s/infra:cat"],
                data = [":%s" % expanded_name],
                args = ["k8s/app/%s" % expanded_name],
            )

def get_k8s_resource_file(dir, cluster_path, environment, dir_prefix = ""):
    """
    Return the kustomization file that applies for the overlay.

    Given dir_prefix "k8s/app" dir "gateway_server" and cluster_path "eu-west/1"
    and environment "dev" it will try

    k8s/app/gateway_server/eu-west/1/dev/kustomization.yaml
    k8s/app/gateway_server/eu-west/1/kustomization.yaml
    k8s/app/gateway_server/eu-west/dev/kustomization.yaml
    k8s/app/gateway_server/eu-west/kustomization.yaml
    k8s/app/gateway_server/dev/kustomization.yaml
    k8s/app/gateway_server/kustomization.yaml
    k8s/app/gateway_server/dev/kustomization.yaml
    k8s/app/gateway_server/kustomization.yaml

    Args:
        dir: Root directory of resource
        cluster_path: The cluster path name
        environment: The cluster environment
        dir_prefix: Prefix directory relative to where deploy_targets is called

    Returns:
        The path (string) to the kustomization.yaml file for the overlay corresponding to the target and cluster.
    """

    base_dir = dir
    if dir_prefix != "":
        base_dir = "%s/%s" % (dir_prefix, dir)

    # print(dir, cluster_path, environment, dir_prefix)
    if cluster_path == "":
        paths = []
    else:
        splits = cluster_path.split("/")
        paths = ["/".join([p for p in splits[:i]]) for i in range(len(splits), 0, -1)]

    for path in paths:
        # Check for a file with cluster and enviroment in path.
        kustomization_files = "%s/%s/%s/kustomization.yaml" % (base_dir, path, environment)

        # print("path: " + kustomization_files)
        if native.glob([kustomization_files], allow_empty = True) != []:
            return kustomization_files

        # Check for a file with cluster in path
        kustomization_files = "%s/%s/kustomization.yaml" % (base_dir, path)

        # print("path: " + kustomization_files)
        if native.glob([kustomization_files], allow_empty = True) != []:
            return kustomization_files

    # Check for a file with enviroment in path.
    kustomization_files = "%s/%s/kustomization.yaml" % (base_dir, environment)

    # print("path: " + kustomization_files)
    if native.glob([kustomization_files], allow_empty = True) != []:
        return kustomization_files

    # Check for a file in root path.
    kustomization_files = "%s/kustomization.yaml" % (base_dir)

    # print("path: " + kustomization_files)
    if native.glob([kustomization_files], allow_empty = True) != []:
        return kustomization_files

    return None
