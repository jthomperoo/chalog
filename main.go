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

	mdrenderer "github.com/Kunde21/markdownfmt/v2/markdown"
	"github.com/Masterminds/semver"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
	"gopkg.in/yaml.v2"
)

// Version specifies the chalog tool version, overridden at build time
var Version string = "development"

const (
	changelogTemplate = `# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).`
	firstDiffTemplate      = "\n[%s]: %s/releases/tag/%s"
	compareDiffTemplate    = "\n[%s]: %s/compare/%s...%s"
	unreleasedDiffTemplate = "\n[%s]: %s/compare/%s...HEAD"
	lineWidth              = 120
	releasesFileName       = "releases.txt"
)

type targetType string

const (
	targetTypeFile   targetType = "file"
	targetTypeStdout targetType = "stdout"
)

const (
	flagIn         = "in"
	flagOut        = "out"
	flagRepo       = "repo"
	flagUnreleased = "unreleased"
	flagConfig     = "config"
	flagTarget     = "target"
	flagVersion    = "version"
)

const (
	defaultIn         = ".changelog"
	defaultOut        = "CHANGELOG.md"
	defaultRepo       = ""
	defaultUnreleased = "Unreleased"
	defaultConfig     = ".chalog.yml"
	defaultTarget     = targetTypeFile
	defaultVersion    = false
)

type config struct {
	In         string     `yaml:"in"`
	Out        string     `yaml:"out"`
	Repo       string     `yaml:"repo"`
	Unreleased string     `yaml:"unreleased"`
	Target     targetType `yaml:"target"`
}

func loadDefaults() *config {
	return &config{
		In:         defaultIn,
		Out:        defaultOut,
		Repo:       defaultRepo,
		Unreleased: defaultUnreleased,
		Target:     defaultTarget,
	}
}

func loadConfig(data []byte) (*config, error) {
	conf := loadDefaults()
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

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func main() {

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: chalog [options]")
		flag.PrintDefaults()
	}

	inFlag := flag.String(flagIn, defaultIn,
		"the directory for storing the changelog files")
	outFlag := flag.String(flagOut, defaultOut,
		"the changelog file to output to")
	repoFlag := flag.String(flagRepo, defaultRepo,
		"the repository base url, include the protocol (http/https)")
	unreleasedFlag := flag.String(flagUnreleased, defaultUnreleased,
		"the release name that should be treated as a the 'unreleased' section")
	configFlag := flag.String(flagConfig, defaultConfig,
		"path to the config file to load")
	targetFlag := flag.String(flagTarget, string(defaultTarget),
		"target to output to, e.g. stdout or a file")
	versionFlag := flag.Bool(flagVersion, defaultVersion,
		"if the process should be skipped, instead printing the version info")

	flag.Parse()

	if *versionFlag {
		fmt.Println(Version)
		return
	}

	configFilePath := *configFlag

	// Read in config file
	configData, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatal(err)
		}
	}

	conf, err := loadConfig(configData)
	if err != nil {
		log.Fatal(err)
	}

	if isFlagPassed(flagIn) {
		conf.In = *inFlag
	}
	if isFlagPassed(flagOut) {
		conf.Out = *outFlag
	}
	if isFlagPassed(flagRepo) {
		conf.Repo = *repoFlag
	}
	if isFlagPassed(flagUnreleased) {
		conf.Unreleased = *unreleasedFlag
	}
	if isFlagPassed(flagTarget) {
		conf.Target = targetType(*targetFlag)
	}

	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithRenderer(mdrenderer.NewRenderer()),
	)

	parser := md.Parser()
	renderer := md.Renderer()

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

			mdRoot := parser.Parse(text.NewReader(dat))
			if err != nil {
				log.Fatal(err)
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
						log.Fatal(err)
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

	textSource := text.NewReader([]byte(changelogTemplate))

	parsedTemplate := parser.Parse(textSource)

	writer := bytes.Buffer{}

	err = renderer.Render(&writer, []byte(changelogTemplate), parsedTemplate)
	if err != nil {
		log.Fatal(err)
	}
	output := writer.String()

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
			output += fmt.Sprintf("%s\n", category)
		}
	}

	if conf.Repo != "" {
		output += "\n"
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

	output += "\n"

	switch conf.Target {
	case targetTypeFile:
		err = ioutil.WriteFile(conf.Out, []byte(output), 0644)
		if err != nil {
			log.Fatal(err)
		}
	case targetTypeStdout:
		fmt.Print(output)
	default:
		log.Fatalf("unknown target type '%s' provided", conf.Target)
	}

}
