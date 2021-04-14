package render

import (
	"fmt"

	"github.com/jthomperoo/chalog/conf"
	"github.com/jthomperoo/chalog/internal/release"
)

const (
	firstDiffTemplate      = "[%s]: %s/releases/tag/%s\n"
	compareDiffTemplate    = "[%s]: %s/compare/%s...%s\n"
	unreleasedDiffTemplate = "[%s]: %s/compare/%s...HEAD\n"
	releasesFileName       = "releases.txt"
)

// Renderer renders releases into a changelog, in the keep a changelog v1.1.0 format
type Renderer struct{}

// NewRenderer sets up a renderer with sensible defaults
func NewRenderer() *Renderer {
	return &Renderer{}
}

// Render takes in some configuration and a list of releases, and renders them into a changelog
func (r *Renderer) Render(config *conf.Config, releases []release.Release) (string, error) {
	output := config.Preamble
	output += r.renderReleases(releases)

	if config.Repo != "" {
		output += r.renderReleaseLinks(config.Unreleased, config.Repo, releases)
	}

	return output, nil
}

func (r *Renderer) renderReleases(releases []release.Release) string {
	output := ""
	// Iterate over releases, splitting release into categories
	for _, release := range releases {
		if release.Meta != "" {
			output += fmt.Sprintf("\n## [%s] - %s\n", release.Name, release.Meta)
		} else {
			output += fmt.Sprintf("\n## [%s]\n", release.Name)
		}
		for categoryName, category := range release.Categories {
			output += fmt.Sprintf("### %s\n", categoryName)
			output += category
		}
	}
	return output
}

func (r *Renderer) renderReleaseLinks(unreleased string, repo string, releases []release.Release) string {
	output := "\n"
	for i, release := range releases {
		if release.Name == unreleased {
			if i+1 >= len(releases) {
				// Only needed to add in link section if there is an actual release, if only unreleased no need
				continue
			}
			output += fmt.Sprintf(unreleasedDiffTemplate, release.Name, repo, releases[i+1].Name)
			continue
		}
		if i == len(releases)-1 {
			// Last one, therefore it's the first release
			output += fmt.Sprintf(firstDiffTemplate, release.Name, repo, release.Name)
			continue
		}
		// Normal, compare with previous
		output += fmt.Sprintf(compareDiffTemplate, release.Name, repo, releases[i+1].Name, release.Name)
	}
	return output
}
