package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	mmark "github.com/mmarkdown/mmark/render/markdown"
	"gopkg.in/yaml.v2"
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
	releasesFileName       = "releases.txt"
)

const (
	defaultIn         = ".changelog"
	defaultOut        = "CHANGELOG.md"
	defaultRepo       = ""
	defaultUnreleased = "Unreleased"
	defaultConfig     = ".chalog.yml"
)

type config struct {
	In         string `yaml:"in"`
	Out        string `yaml:"out"`
	Repo       string `yaml:"repo"`
	Unreleased string `yaml:"unreleased"`
}

func loadConfig(data []byte, conf *config) (*config, error) {
	err := yaml.Unmarshal([]byte(data), conf)
	if err != nil {
		return nil, err
	}

	return conf, nil
}

type release struct {
	name       string
	meta       string
	categories map[string]string
}

func main() {

	inFlag := flag.String("in", defaultIn,
		"the directory for storing the changelog files")
	outFlag := flag.String("out", defaultOut,
		"the changelog file to output to")
	repoFlag := flag.String("repo", defaultRepo,
		"the repository base url, include the protocol (http/https etc.)")
	unreleasedFlag := flag.String("unreleased", defaultUnreleased,
		"the release name that should be treated as a the 'unreleased' section")
	configFlag := flag.String("config", defaultConfig,
		"the optional path to the config file to load")

	flag.Parse()

	inOpt := *inFlag
	outOpt := *outFlag
	repoOpt := *repoFlag
	unreleasedOpt := *unreleasedFlag
	configOpt := *configFlag

	markdownRenderer := mmark.NewRenderer(mmark.RendererOptions{
		TextWidth: lineWidth,
		Flags:     mmark.CommonFlags,
	})

	// Read in config file
	configData, err := ioutil.ReadFile(configOpt)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatal(err)
		}
	}

	conf, err := loadConfig(configData, &config{
		In:         inOpt,
		Out:        outOpt,
		Repo:       repoOpt,
		Unreleased: unreleasedOpt,
	})
	if err != nil {
		log.Fatal(err)
	}

	releases := []release{}

	// First check for releases.txt, if it exists read it in
	providedReleasesFile := true
	releaseFileDat, err := ioutil.ReadFile(filepath.Join(conf.In, releasesFileName))
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatal(err)
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

	releaseDirectories, err := ioutil.ReadDir(conf.In)
	if err != nil {
		log.Fatal(err)
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

		changeFiles, err := ioutil.ReadDir(filepath.Join(conf.In, releaseName))
		if err != nil {
			log.Fatal(err)
		}
		for _, changeFile := range changeFiles {
			if changeFile.IsDir() {
				continue
			}

			dat, err := ioutil.ReadFile(filepath.Join(conf.In, releaseName, changeFile.Name()))
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
						newRelease.categories[currentCategory] = categoryText
					}
					currentCategory = headingTitle
					categoryText = newRelease.categories[currentCategory]
				default:
					categoryText += string(markdown.Render(child, markdownRenderer))
				}
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

	parsedTemplate := markdown.Parse([]byte(changelogTemplate), nil)

	output := string(markdown.Render(parsedTemplate, markdownRenderer))

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
			output += fmt.Sprintf("\n## [%s] - %s\n", release.name, release.meta)
		} else {
			output += fmt.Sprintf("\n## [%s]\n", release.name)
		}
		for categoryName, category := range release.categories {
			output += fmt.Sprintf("### %s\n", categoryName)
			output += category
		}
	}

	if conf.Repo != "" {
		// Create diffs for releases
		for i, release := range releases {
			if release.name == conf.Unreleased {
				if i+1 >= len(releases) {
					// Only needed to add in diff section if there is an actual release, if only unreleased no need
					continue
				}
				output += fmt.Sprintf(unreleasedDiffTemplate, release.name, conf.Repo, releases[i+1].name)
				continue
			}
			if i == len(releases)-1 {
				// Last one, therefore it's the first release
				output += fmt.Sprintf(firstDiffTemplate, release.name, conf.Repo, release.name)
				continue
			}
			// Normal, compare with previous
			output += fmt.Sprintf(compareDiffTemplate, release.name, conf.Repo, releases[i+1].name, release.name)
		}
	}

	err = ioutil.WriteFile(conf.Out, []byte(output), 0644)
}
