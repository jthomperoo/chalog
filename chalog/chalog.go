package chalog

import (
	"github.com/jthomperoo/chalog/conf"
	"github.com/jthomperoo/chalog/internal/generator"
)

// Generator is a chalog generator, taking configuration and outputting a changelog
type Generator interface {
	Generate(config *conf.Config) (string, error)
}

// NewGenerator creates a chalog generator set up with sensible defaults
func NewGenerator() Generator {
	return generator.NewGenerator()
}
