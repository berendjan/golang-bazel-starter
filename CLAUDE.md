# Golang-bazel-starter

When changing dependencies in go code do the following:
After adding a new golang library:

```bash
# update go.mod
go mod tidy -e

# update bazel go deps
bazel mod tidy

# update BUILD files
bazel run //:gazelle

# build target with bazel
```

Do not manually update BUILD.bazel files, always do
```bash
bazel run //:gazelle
```



