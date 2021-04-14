package generator

import (
	"github.com/jthomperoo/chalog/conf"
	"github.com/jthomperoo/chalog/internal/parse"
	"github.com/jthomperoo/chalog/internal/release"
	"github.com/jthomperoo/chalog/internal/render"
)

// Parser takes a configuration and parses releases from it
type Parser interface {
	Parse(config *conf.Config) ([]release.Release, error)
}

// Renderer takes a list of releases and some configuration and renders a changelog from it
type Renderer interface {
	Render(config *conf.Config, releases []release.Release) (string, error)
}

// Generator generates a changelog from a directory structured in the chalog format
type Generator struct {
	Parser   Parser
	Renderer Renderer
}

// NewGenerator sets up a generator with sensible defaults
func NewGenerator() *Generator {
	return &Generator{
		Parser:   parse.NewParser(),
		Renderer: render.NewRenderer(),
	}
}

// Generate takes a config and uses it to generate a chalog changelog
func (g *Generator) Generate(config *conf.Config) (string, error) {
	releases, err := g.Parser.Parse(config)
	if err != nil {
		return "", err
	}
	return g.Renderer.Render(config, releases)
}
