# Bazel Golang Starter

## Dev commands

```bash
# update go.mod
go mod tidy
# update golang modules
bazel mod tidy
# update BUILD.bazel files
bazel run //:gazelle
