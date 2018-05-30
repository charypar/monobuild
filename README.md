# Monobuild

A build orchestration tool for Continuous Integration in a monorepo.

NOTE: this is Readme driven development. Not everything described in this readme
is fully implemented.

## About

Monobuild is a simple tool that understands a graph of dependencies in
a monorepo codebase (where separate components live side by side in folders)
and based on it, it can decide what should be built, given a set of changes.

For help, run

```sh
$ monobuild help
```

## Usage

Monobuild constructs the dependency graph from dependency manifests. By
default, manifests are files named `Dependencies`, which contain a simple
line-by-line list of dependencies for the component in the directory of the
file.

### Declare dependencies

An example manifest in `app1/Dependencies` might look like this

```
# Content
!data/content
!shared/images

# Libs
common-lib
libs/date-time
```

Monobuild will ignore any empty lines and lines starting with `#`. Every other
line is considered a dependency and is a path relative to current working
directory (typically repository root). Monobuild will expect a dependency
manifest (possibly empty) to be present at that path.

Lines starting with `!` are strong dependencies all other dependencies are
considered weak. The difference is in the way the dependency graph is translated
to a build schedule.

One of the benefits of a monorepo, is components and services can be built from
code, including their dependencies. Changing a weak dependency of a component
means a change to the component, which therefore needs to be rebuilt, but the
builds can be run in parallel. Output or result of the dependency does not
affect the build of this component.

A strong dependency has to successfully build first, in order for the build of
the component to be possible. If the dependency build fails, the component
build does not even start.

Typically, services are built from source, including their libraries, so the
dependencies on libraries are weak (we still want to run the library build to
run tests and get a result though). Deploying orchestrations of services
typically has a strong dependency on the service builds (as they produce
artifacts, e.g. docker images, needed by the deployment).

### Visualise dependency graph and build schedule

To better understand the dependency graphs and build schedules, Monobuild can
print them.

```sh
$ monobuild print
```

will print the build schedule, which will ignore weak dependencies

```sh
$ cd test/fixtures/manifests-test
$ monobuild print
app4:
libs/lib1:
libs/lib2:
libs/lib3:
stack1: app1, app2, app3
app1:
app2:
app3:
```

You can also print the dependency structure with one component per line. For example

```sh
$ cd test/fixtures/manifests-test
$ monobuild print --dependencies
app4:
libs/lib1: libs/lib3
libs/lib2: libs/lib3
libs/lib3:
stack1: app1, app2, app3
app1: libs/lib1, libs/lib2
app2: libs/lib2, libs/lib3
app3: libs/lib3
```

Print also supports graphical output using GraphViz

```
$ cd test/fixtures/manifests-test
$ monobuild print --dot
```

to produce a PDF, you can pipe the output into the `dot` tool:

```
$ cd test/fixtures/manifests-test
$ monobuild print --dependencies --dot | dot -Tpdf -o dependencies.pdf
digraph dependencies {
  "app1" -> "libs/lib1"
  "app1" -> "libs/lib2"
  "app2" -> "libs/lib2"
  "app2" -> "libs/lib3"
  "app3" -> "libs/lib3"
  "libs/lib1" -> "libs/lib3"
  "libs/lib2" -> "libs/lib3"
  "stack1" -> "app1"
  "stack1" -> "app2"
  "stack1" -> "app3"
}
```

### Change detection

If the current directory is a git repository, monobuild can decide which
components changed (using git).

```sh
$ monobuild diff
app2
app2
lib3
```

Monobuild assumes use of [Mainline Development](https://gitversion.readthedocs.io/en/latest/reference/mainline-development/)
and changes are detected in two modes:

1.  for a feature branch, the change detection is equivalent to

    ```sh
    $ git diff --no-commit-id --name-only -r $(git merge-base master HEAD)
    ```

    in other words, list all the changes that happened since the current branch
    was cut from `master`.

    This is the default mode and the base branch is `master` by default.
    You can override this with

    ```sh
    $ monobuild diff --base-branch develop
    ```

2.  for a `master` branch (or other main branch) the change detection is equivalent
    to

    ```sh
    $ git diff --no-commit-id --name-only -r HEAD^1
    ```

    To work in the main-branch mode, use the `--main-branch` flag

    ```sh
    $ monobuild diff --main-branch
    ```

The main difference between the above `git diff`s and `monobuild diff` is the
dependency graph awareness.

Monobuild will start with the list from `git diff`, filter it down to known
components, and then extend it with all components that depend on any of the
components in the initial list, including transitive dependencies.

For the resulting "to do" list, `diff` will then build a build schedule using the
strong dependencies.

You can print the relevant part of the dependency graph (rather than
the build schedule) with `--dependencies`

```
$ monobuild diff --dependencies
```

Both modes also support DOT output with `--dot`. You can also print
the entire graph with the affected components with `--dot-highlight`.

#### Rebuilding strong dependencies

The assumption behind strong dependencies is that their outcome is required
for the dependent builds to proceed. In most cases, this means that if no
changes affected a component, the build does not need to run, because the outcome
(e.g. a build artifact) already exists from a previous run of the build (when
that component was affected).

In certain situations, it could be useful to run the build again, to ensure its
output is present. This will result in wasted work, but ensures builds won't
fail because, for example, an artifact cache was lost. The wasted work can
also largely be prevented by making builds idempotent.

Monobuild supports this with an `--rebuild-strong` option on `diff`, which will
include strong dependencies of all components affected by the change.

### Override the manifest matching

If you want to use a different filename for the manifest files, you can do so
using the global `--manifests` flag.

### Filters

#### Scope

You can scope the results of both `diff` and `print` to a given component
and its dependencies using the `--scope` flag

#### Top-level components

Sometimes it's useful to know the "entrypoints" into your dependency graph -
the components that nothing depends on (typically services or applications).
You can list only those with a `--top-level` flag on both `diff` and `print`.

### Creating a Makefile

Monobuild can also generate a `Makefile`, that can be used by individual
component builds to build their dependencies.

You can generate the makefile with

```sh
$ monobuild makefile
```

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

You can also override the build command (`make build` by default) with the
`--build-command` flag.
