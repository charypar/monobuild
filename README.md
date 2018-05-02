# Monobuild

A build orchestration tool for Continuous Integration in a monorepo.

NOTE: this is Readme driven development. Not everything described in this readme
is fully implemented.

## About

Monobuild is a simple tool that understands a graph of dependencies in
a monorepo codebase (where separate components live side by side) and
based on it can decide what should be built, given a set of changes.

For help, run

```sh
$ mb help
```

It can do three basic things

### Change detection

If the current directory is a git repository, given a commit sha,
monobuild can decide which components changed (using git).

```sh
$ mb diff
app2
app2
lib3
```

To decide which paths are a component, `monobuild` looks for files
named `Dependencies`, which contain a simple line-by-line list of
dependencies for a given component. Each line is an asbolute path
in the repository (and there should be a `Dependencies` file on that path).

You can override the search pattern with the `--dependency-files`. The pattern
supports the same glob rules as the `find` command.

```sh
$ mb diff --dependency-files **/Dependencies
```

Monobuild assumes use of [Mainline Development]() and changes are detected
in two modes:

1.  for a feature branch, the change detection is equivalent to

    ```sh
    $ git diff --no-commit-id --name-only -r $(git merge-base origin/master HEAD)
    ```

    in other words, list all the changes that happened since the current branch
    was cut from `origin/master`.

    This is the default mode and the base branch is `origin/master` by default.
    You can override this with

    ```sh
    $ mb diff --base-branch origin/master
    ```

2.  for a master branch (or other main branch) the change detection is equivalent
    to

    ```sh
    $ git diff --no-commit-id --name-only -r HEAD^1
    ```

    To work in the main-branch mode, use the `--main-branch` flag

    ```sh
    $ mb diff --main-branch
    ```

### Walking the dependency graph

The main difference between the above `git diff`s and `mb diff` is the
dependency graph awareness.

Monobuild will start with the list from git diff, filter it down to known
components, and then extend it with all components that depend on any of the
components in the initial list, including transitive dependencies. The dependency
graph is built from the `Dependencies` files.

### Creating a Makefile

Monobuild can also generate a `Makefile`, that can be used by individual
component builds to assemble themselves. The make dependency graph is
the same graph as the one used to decide what should build (except the edges
are reversed).

You can generate the makefile with

```sh
$ mb makefile
```

The command supports the `--dependency-files` flag the same way `diff` does.

The resulting Makefile consists of targets like this:

```make
directory/component1: [dependency1] [dependency2] [dependency3]
  @cd directory/component1 && make build
```

This assumes each component has a minimal `Makefile` which looks like this:

```make
# directory/component1/Makefile

default:
  @cd ../.. && make directory/component1

build:
  # steps to make the component available as a dependency of others
  # this could be empty
```
