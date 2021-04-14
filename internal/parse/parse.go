package parse

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/jthomperoo/chalog/conf"
	"github.com/jthomperoo/chalog/internal/release"
)

const (
	releasesFileName = "releases.txt"
)

// Parser uses a chalog configuration to parse out releases from the directory structure provided.
type Parser struct{}

// NewParser sets up a parser with sensible defaults
func NewParser() *Parser {
	return &Parser{}
}

// Parse takes in a configuration and uses it to query a directory and its subdirectories to build up a sorted list
// of releases
func (p *Parser) Parse(config *conf.Config) ([]release.Release, error) {
	releases := []release.Release{}

	// First get releases file if it exists
	hasReleasesFile, releaseFile, err := p.getReleaseFile(filepath.Join(config.In, releasesFileName))
	if err != nil {
		return nil, err
	}

	// If it exists parse any releases from it
	if hasReleasesFile {
		releaseList, err := p.parseReleaseFile(bufio.NewReader(releaseFile))
		if err != nil {
			return nil, err
		}
		releases = append(releases, releaseList...)
	}

	// Parse the releases directory, splitting sub directories into releases
	releases, err = p.parseReleaseDirectory(config.In, releases)
	if err != nil {
		return nil, err
	}

	if !hasReleasesFile {
		// Sort releases by semantic versioning, if the release names are not valid semantic versions they will be
		// sorted alphabetically, if there is a mix of semvers and not semvers, the non semvers will be placed first
		sort.Slice(releases, func(i, j int) bool {
			iV, iErr := semver.NewVersion(releases[i].Name)
			jV, jErr := semver.NewVersion(releases[j].Name)
			if iErr != nil || jErr != nil {
				// If i or j is not a semantic version
				if iErr != nil && jErr != nil {
					// If both are not not semantic versions, just compare alphabetically
					return releases[i].Name < releases[j].Name
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

	return releases, nil
}

func (p *Parser) getReleaseFile(path string) (bool, *os.File, error) {
	file, err := os.Open(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return false, nil, err
		}
		// file doesn't exist
		return false, nil, nil
	}
	return true, file, nil
}

func (p *Parser) parseReleaseFile(r io.Reader) ([]release.Release, error) {
	releases := []release.Release{}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		parts := strings.Split(line, ",")

		name := parts[0]

		alreadySet := false
		for _, release := range releases {
			if release.Name == name {
				alreadySet = true
				break
			}
		}

		if alreadySet {
			continue
		}

		if len(parts) < 2 {
			releases = append(releases, release.Release{
				Name: name,
			})
			continue
		}

		meta := parts[1]

		releases = append(releases, release.Release{
			Name: name,
			Meta: meta,
		})
	}
	return releases, nil
}

func (p *Parser) parseReleaseDirectory(in string, releases []release.Release) ([]release.Release, error) {
	releaseDirectories, err := ioutil.ReadDir(in)
	if err != nil {
		return nil, err
	}

	for _, releaseDir := range releaseDirectories {
		if !releaseDir.IsDir() {
			continue
		}
		releaseName := releaseDir.Name()

		newRelease := release.Release{
			Name:       releaseName,
			Categories: make(map[string]string),
		}

		changeFiles, err := ioutil.ReadDir(filepath.Join(in, releaseName))
		if err != nil {
			return nil, err
		}
		for _, changeFile := range changeFiles {
			if changeFile.IsDir() {
				continue
			}

			file, err := os.Open(filepath.Join(in, releaseName, changeFile.Name()))
			if err != nil {
				return nil, err
			}

			var currentCategory string
			var categoryText string

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				trimmedLine := strings.TrimSpace(line)
				if len(trimmedLine) > 1 && trimmedLine[0:2] == "# " {
					// If the line is a level 1 header, set up as a new category
					category := strings.TrimSpace(trimmedLine[1:])
					if currentCategory != "" {
						newRelease.Categories[currentCategory] = categoryText
					}
					currentCategory = category
					categoryText = newRelease.Categories[currentCategory]
				} else if trimmedLine != "" {
					// Ignore lines that are just spaces
					// Otherwise, just add the line to the category text
					categoryText += line + "\n"
				}
			}

			if currentCategory != "" {
				newRelease.Categories[currentCategory] = categoryText
			}
		}

		alreadySet := false

		for i, release := range releases {
			if release.Name == newRelease.Name {
				releases[i].Categories = newRelease.Categories
				alreadySet = true
				break
			}
		}

		if !alreadySet {
			releases = append(releases, newRelease)
		}
	}
	return releases, nil
}
