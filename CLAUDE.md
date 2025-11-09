# Golang-bazel-starter

When changing dependencies in go code do the following

```bash
# update go.mod
go mod tidy -e

# update bazel go deps
bazel mod tidy

# update BUILD files
bazel run //:gazelle

# build target with bazel
```

