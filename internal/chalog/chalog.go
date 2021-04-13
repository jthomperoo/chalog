package chalog

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/jthomperoo/chalog/conf"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

const (
	changelogTemplate = `# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).`
	firstDiffTemplate      = "[%s]: %s/releases/tag/%s\n"
	compareDiffTemplate    = "[%s]: %s/compare/%s...%s\n"
	unreleasedDiffTemplate = "[%s]: %s/compare/%s...HEAD\n"
	releasesFileName       = "releases.txt"
)

type release struct {
	name       string
	meta       string
	categories map[string]string
}

// Generator generates a changelog from a directory structured in the chalog format
type Generator struct {
	Markdown goldmark.Markdown
}

// Generate takes a config and uses it to generate a chalog changelog
func (g *Generator) Generate(config *conf.Config) (string, error) {
	parser := g.Markdown.Parser()
	renderer := g.Markdown.Renderer()

	releases := []release{}

	// First check for releases.txt, if it exists read it in
	providedReleasesFile := true
	releaseFileDat, err := ioutil.ReadFile(filepath.Join(config.In, releasesFileName))
	if err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}
		// file doesn't exist
		providedReleasesFile = false
	}

	if providedReleasesFile {
		scanner := bufio.NewScanner(bytes.NewReader(releaseFileDat))
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}

			parts := strings.Split(line, ",")

			name := parts[0]

			if len(parts) < 2 {
				releases = append(releases, release{
					name: name,
				})
				continue
			}

			meta := parts[1]

			releases = append(releases, release{
				name: name,
				meta: meta,
			})
		}
	}

	releaseDirectories, err := ioutil.ReadDir(config.In)
	if err != nil {
		return "", err
	}

	for _, releaseDir := range releaseDirectories {
		if !releaseDir.IsDir() {
			continue
		}
		releaseName := releaseDir.Name()

		newRelease := release{
			name:       releaseName,
			categories: make(map[string]string),
		}

		changeFiles, err := ioutil.ReadDir(filepath.Join(config.In, releaseName))
		if err != nil {
			return "", err
		}
		for _, changeFile := range changeFiles {
			if changeFile.IsDir() {
				continue
			}

			dat, err := ioutil.ReadFile(filepath.Join(config.In, releaseName, changeFile.Name()))
			if err != nil {
				return "", err
			}

			mdRoot := parser.Parse(text.NewReader(dat))
			if err != nil {
				return "", err
			}

			var currentCategory string
			var categoryText string

			head := mdRoot.FirstChild()

			for head != nil {
				kind := head.Kind()

				if kind.String() == "Heading" {
					heading := head.(*ast.Heading)
					headingText := heading.BaseBlock.FirstChild().(*ast.Text)
					headingTitle := string(headingText.Segment.Value(dat))
					if currentCategory != "" {
						newRelease.categories[currentCategory] = categoryText
					}
					currentCategory = headingTitle
					categoryText = newRelease.categories[currentCategory]
				} else {
					writer := bytes.Buffer{}
					err = renderer.Render(&writer, dat, head)
					if err != nil {
						return "", err
					}
					categoryText += strings.TrimPrefix(writer.String(), "\n")
				}

				head = head.NextSibling()
			}

			if currentCategory != "" {
				newRelease.categories[currentCategory] = categoryText
			}

		}

		alreadySet := false

		for i, release := range releases {
			if release.name == newRelease.name {
				releases[i].categories = newRelease.categories
				alreadySet = true
				break
			}
		}

		if !alreadySet {
			releases = append(releases, newRelease)
		}
	}

	output := changelogTemplate

	if !providedReleasesFile {
		// Sort releases by semantic versioning, if the release names are not valid semantic versions they will be
		// sorted alphabetically, if there is a mix of semvers and not semvers, the non semvers will be placed first
		sort.Slice(releases, func(i, j int) bool {
			iV, iErr := semver.NewVersion(releases[i].name)
			jV, jErr := semver.NewVersion(releases[j].name)
			if iErr != nil || jErr != nil {
				// If i or j is not a semantic version
				if iErr != nil && jErr != nil {
					// If both are not not semantic versions, just compare alphabetically
					return releases[i].name < releases[j].name
				}
				if iErr != nil {
					// If just i is not a semantic version
					return true
				}
				// If just j is not a semantic version
				return false
			}
			return iV.GreaterThan(jV)
		})
	}

	// Create releases section
	for _, release := range releases {
		if release.meta != "" {
			output += fmt.Sprintf("\n\n## [%s] - %s\n", release.name, release.meta)
		} else {
			output += fmt.Sprintf("\n\n## [%s]\n", release.name)
		}
		for categoryName, category := range release.categories {
			output += fmt.Sprintf("### %s\n", categoryName)
			output += strings.TrimPrefix(category, "\n")
		}
	}

	if config.Repo != "" {
		output += "\n\n"
		// Create diffs for releases
		for i, release := range releases {
			if release.name == config.Unreleased {
				if i+1 >= len(releases) {
					// Only needed to add in diff section if there is an actual release, if only unreleased no need
					continue
				}
				output += fmt.Sprintf(unreleasedDiffTemplate, release.name, config.Repo, releases[i+1].name)
				continue
			}
			if i == len(releases)-1 {
				// Last one, therefore it's the first release
				output += fmt.Sprintf(firstDiffTemplate, release.name, config.Repo, release.name)
				continue
			}
			// Normal, compare with previous
			output += fmt.Sprintf(compareDiffTemplate, release.name, config.Repo, releases[i+1].name, release.name)
		}
	}

	output = strings.TrimSuffix(output, "\n")
	output += "\n"
	return output, nil
}
