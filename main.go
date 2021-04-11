package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"

	"github.com/Masterminds/semver"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	mmark "github.com/mmarkdown/mmark/render/markdown"
)

const (
	changelogTemplate = `# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
`
	firstDiffTemplate      = "\n[%s]: %s/releases/tag/%s"
	compareDiffTemplate    = "\n[%s]: %s/compare/%s...%s"
	unreleasedDiffTemplate = "\n[%s]: %s/compare/%s...HEAD"
	lineWidth              = 120
)

const (
	defaultChangelogDir      = ".changelog"
	defaultChangelogFile     = "CHANGELOG.md"
	defaultRepositoryBaseURL = ""
	defaultUnreleasedName    = "Unreleased"
)

type release struct {
	name       string
	categories map[string]string
}

func main() {

	changelogDirFlag := flag.String("in", defaultChangelogDir,
		"the directory for storing the changelog files")
	changelogFileFlag := flag.String("out", defaultChangelogFile,
		"the changelog file to output to")
	repositoryBaseURLFlag := flag.String("repo", defaultRepositoryBaseURL,
		"the repository base url, include the protocol (http/https etc.)")
	unreleasedNameFlag := flag.String("unreleased", defaultUnreleasedName,
		"the release name that should be treated as a the 'unreleased' section")

	flag.Parse()

	changelogDir := *changelogDirFlag
	changelogFile := *changelogFileFlag
	repositoryBaseURL := *repositoryBaseURLFlag
	unreleasedName := *unreleasedNameFlag

	markdownRenderer := mmark.NewRenderer(mmark.RendererOptions{
		TextWidth: lineWidth,
		Flags:     mmark.CommonFlags,
	})

	releaseDirectories, err := ioutil.ReadDir(changelogDir)
	if err != nil {
		log.Fatal(err)
	}

	releases := []release{}
	for _, releaseDir := range releaseDirectories {
		if !releaseDir.IsDir() {
			continue
		}
		releaseName := releaseDir.Name()
		release := release{
			name:       releaseName,
			categories: make(map[string]string),
		}

		changeFiles, err := ioutil.ReadDir(filepath.Join(changelogDir, releaseName))
		if err != nil {
			log.Fatal(err)
		}
		for _, changeFile := range changeFiles {
			if changeFile.IsDir() {
				continue
			}

			dat, err := ioutil.ReadFile(filepath.Join(changelogDir, releaseName, changeFile.Name()))
			if err != nil {
				log.Fatal(err)
			}

			mdRoot := markdown.Parse(dat, nil)
			if err != nil {
				log.Fatal(err)
			}

			children := mdRoot.GetChildren()

			var currentCategory string
			var categoryText string

			for _, child := range children {
				switch v := child.(type) {
				case *ast.Heading:
					headingTitle := string(v.GetChildren()[0].AsLeaf().Literal)
					if currentCategory != "" {
						release.categories[currentCategory] = categoryText
					}
					currentCategory = headingTitle
					categoryText = release.categories[currentCategory]
				default:
					categoryText += string(markdown.Render(child, markdownRenderer))
				}
			}

			if currentCategory != "" {
				release.categories[currentCategory] = categoryText
			}

		}

		releases = append(releases, release)
	}

	parsedTemplate := markdown.Parse([]byte(changelogTemplate), nil)

	output := string(markdown.Render(parsedTemplate, markdownRenderer))

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

	// Create releases section
	for _, release := range releases {
		output += fmt.Sprintf("\n## [%s]\n", release.name)
		for categoryName, category := range release.categories {
			output += fmt.Sprintf("### %s\n", categoryName)
			output += category
		}
	}

	if repositoryBaseURL != "" {
		// Create diffs for releases
		for i, release := range releases {
			if release.name == unreleasedName {
				if len(releases) < 2 {
					// Only needed to add in diff section if there is an actual release, if only unreleased no need
					continue
				}
				output += fmt.Sprintf(unreleasedDiffTemplate, release.name, repositoryBaseURL, releases[i+1].name)
				continue
			}
			if i == len(releases)-1 {
				// Last one, therefore it's the first release
				output += fmt.Sprintf(firstDiffTemplate, release.name, repositoryBaseURL, release.name)
				continue
			}
			// Normal, compare with previous
			output += fmt.Sprintf(compareDiffTemplate, release.name, repositoryBaseURL, releases[i+1].name, release.name)
		}
	}

	err = ioutil.WriteFile(changelogFile, []byte(output), 0644)
}
