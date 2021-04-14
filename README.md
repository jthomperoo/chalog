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

## Installation

If using a Go development environment, chalog can be installed by running:

```bash
go install -ldflags="-X 'main.Version=v0.3.0'" github.com/jthomperoo/chalog@v0.3.0
```

Otherwise, the packaged binaries can be used, [check out the available binaries for `v0.3.0` from the GitHub releases
page here](https://github.com/jthomperoo/chalog/releases/tag/v0.3.0).

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

You can see the latest generated changelog [attached as an asset to the latest
release](https://github.com/jthomperoo/chalog/releases/tag/v0.3.0).

## Configuration

The chalog tool includes some configuration options that can be set by flags provided to the executable. These are all
optional.

```
Usage: chalog [options]
  -config string
    	path to the config file to load (default ".chalog.yml")
  -in string
    	the directory for storing the changelog files (default ".changelog")
  -out string
    	the changelog file to output to (default "CHANGELOG.md")
  -repo string
    	the repository base url, include the protocol (http/https)
  -target string
    	target to output to, e.g. stdout or a file (default "file")
  -unreleased string
    	the release name that should be treated as a the 'unreleased' section (default "Unreleased")
  -version
    	if the process should be skipped, instead printing the version info
```

### Releases file

The chalog tool allows you to provide an optional `releases.txt` file inside your changelog directory, which allows
you to provide an ordering for the releases in the changelog, and to optionally provide metadata (e.g. the date of a
release). If this file does not exist then the releases will just be sorted according to semantic versioning (as best
as possible) and without any metadata.

To use the releases file, create a `releases.txt` in the changelog directory with the following format:

```txt
v0.2.0,2021-04-01
v0.1.0,2021-03-07
```

This will result in a generated changelog which will follow this order, with the dates provided appended as metadata
to the release:

```markdown
## [v0.1.0] - 2021-03-07
```

If a release is mentioned in the releases file, but there is no directory, the release will be added to the changelog
with no changes listed.

### Config file

The chalog tool allows you to provide an optional YAML configuration, which lets you set any option that is avaialble
as a command line option (other than `config`) through a YAML file, which can be easily stored in source control.

The default location of the configuration file is `.chalog.yml` in the directory chalog is run in, if there is no
file provided it will ignore it and continue as normal. A path to the configuration file can be provided with the
`-config` command line option, e.g. `-config my_chalog_config.yml`.

The YAML configuration looks like this:

```yaml
in: .changelog
out: CHANGELOG.md
unreleased: Unreleased
repo: https://github.com/jthomperoo/chalog
```

The configuration file is overridden by any command line options provided, so can act as sensible defaults that
can be modified by adjusting command line options provided.

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
