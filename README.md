[![Build](https://github.com/jthomperoo/chalog/workflows/main/badge.svg)](https://github.com/jthomperoo/chalog/actions)
[![go.dev](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/jthomperoo/chalog)
[![Go Report
Card](https://goreportcard.com/badge/github.com/jthomperoo/chalog)](https://goreportcard.com/report/github.com/jthomperoo/chalog)
[![License](https://img.shields.io/:license-mit-blue.svg)](https://choosealicense.com/licenses/mit/)

# chalog

This is chalog, a changelog management tool. With chalog you can manage your project's changelog in a simple markdown
format, split across multiple files, and then use chalog to generate the final combined changelog.

This is primarily to address the issue of merge conflicts in changelogs, which can occur easily if there are multiple
people working on a single project (there have been a [few articles with other
solutions](https://about.gitlab.com/blog/2018/07/03/solving-gitlabs-changelog-conflict-crisis/) to this problem, this
is just a different solution).

## Format

The chalog tool currently only works with the [keep a changelog v1.1.0](https://keepachangelog.com/en/1.1.0/) changelog
format, and on releases hosted on GitHub.

The format for chalog is intended to be simple, readable, and not require much knowledge or configuration - chalog
works out of the box without any configuration.

The format for storing your changelog is a changelog directory with subdirectories, with each subdirectory being
a release in your changelog. Inside these release directories there can be any number of markdown files (no markdown
files is still valid) which contain the changes for the release. When the chalog tool is run these files are read
and grouped by release, combining them into a single changelog markdown file.

It is important that when you create your markdown files as part of your changelog, they should be split up by
top level headers (e.g. `# Added`) because chalog uses these headers to group changes. If no header is provided, the
file will be ignored.

## Quick start

1. Install chalog.

2. Add a new `.changelog` directory to the root directory of your project.

3. Add a new `Unreleased` subdirectory to your new `.changelog` directory.

4. Add a new `v1.0.0` subdirectory to your new `.changelog` directory.

5. Create a new markdown file under the `v1.0.0` subdirectory called `init.md` with the following content:

```md
# Added
- Initial release of my project!
```

6. Now run `chalog` in the root directory of your project, if it is successful it will generate a `CHANGELOG.md` file
that will contain your new combined changelog.

7. You now have a chalog formatted changelog set up for your project, try adding some more markdown files under
releases, or adding new releases and check how the `CHANGELOG.md` file is generated after running `chalog` again.

## Examples

For an example of a project using chalog, look at the chalog project itself - it uses itself to manage its changelog.

## Configuration

The chalog tool includes some configuration options that can be set by flags provided to the executable. These are all
optional.

```
Usage of chalog:
  -in string
    	the directory for storing the changelog files (default ".changelog")
  -out string
    	the changelog file to output to (default "CHANGELOG.md")
  -repo string
    	the repository base url, include the protocol (http/https etc.)
  -unreleased string
    	the release name that should be treated as a the 'unreleased' section (default "Unreleased")
```

## Contributing

Feel free to contribute to this project, here's some useful info for getting set up.

### Dependencies

Developing this project requires these dependencies:

* [Go](https://golang.org/doc/install) >= `1.16`
* [Golint](https://github.com/golang/lint) == `v0.0.0-20201208152925-83fdc39ff7b5`

### Commands

Use the following commands to develop:

- `go run main.go` = run the tool.
- `make lint` = run the code linter against the code, if this does not pass the CI will not pass.
- `make beautify` = run the code beautifier against the code, this must be run for the CI to pass.
- `make test` = run the project tests.
- `make` = cross compile all supported architectures and operating systems.
- `make zip` = zip up all the compiled and built binaries.
