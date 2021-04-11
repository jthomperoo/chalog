# Added

- Can generate a `CHANGELOG.md` file from a directory of markdown files.
- Can provide some flags for customising the output.
  - `in` = set the changelog directory to read (default `.changelog`).
  - `out` = set the changelog file output to generate (default `CHANGELOG.md`).
  - `repo` = set the project's base repository, used to generate links for comparing releases (default empty).
  - `unreleased` = set the changelog's 'unreleased' section name, meaning that any entries under this section will
  be treated as unreleased (default `Unreleased`).
