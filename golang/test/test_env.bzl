"""Custom go_test wrapper that sets testcontainers environment variables."""

load("@rules_go//go:def.bzl", _go_test = "go_test")

def go_test(name, env = None, **kwargs):
    """Wrapper around go_test that automatically disables Ryuk for testcontainers.

    Args:
        name: The name of the test target
        env: Environment variables (will be merged with testcontainers settings)
        **kwargs: All other arguments passed to go_test
    """
    # Start with default testcontainers environment
    test_env = {
        "TESTCONTAINERS_RYUK_DISABLED": "true",
    }

    # Merge with any user-provided env variables
    if env:
        test_env.update(env)

    # Call the original go_test with merged env
    _go_test(
        name = name,
        env = test_env,
        **kwargs
    )
