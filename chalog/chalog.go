package chalog

import (
	"github.com/jthomperoo/chalog/conf"
	"github.com/jthomperoo/chalog/internal/chalog"
	"github.com/jthomperoo/chalog/internal/renderer"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

// Generator is a chalog generator, taking configuration and outputting a changelog
type Generator interface {
	Generate(config *conf.Config) (string, error)
}

// NewGenerator creates a chalog generator set up with sensible defaults
func NewGenerator() Generator {
	return &chalog.Generator{
		Markdown: goldmark.New(
			goldmark.WithExtensions(extension.GFM),
			goldmark.WithRenderer(renderer.NewRenderer()),
		),
	}
}
