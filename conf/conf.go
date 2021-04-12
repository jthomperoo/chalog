package conf

// TargetType is the method for outputting the result of the chalog changelog generation
type TargetType string

const (
	// TargetTypeFile is the target type for outputting to a file
	TargetTypeFile TargetType = "file"
	// TargetTypeStdout is the target type for outputting to stdout
	TargetTypeStdout TargetType = "stdout"
)

const (
	// DefaultIn is the default value for the 'in' option
	DefaultIn = ".changelog"
	// DefaultOut is the default value for the 'out' option
	DefaultOut = "CHANGELOG.md"
	// DefaultRepo is the default value for the 'repo' option
	DefaultRepo = ""
	// DefaultUnreleased is the default value for the 'unreleased' option
	DefaultUnreleased = "Unreleased"
	// DefaultTarget is the default value for the 'target' option
	DefaultTarget = TargetTypeFile
)

// Config is the chalog configuration options
type Config struct {
	In         string     `yaml:"in"`
	Out        string     `yaml:"out"`
	Repo       string     `yaml:"repo"`
	Unreleased string     `yaml:"unreleased"`
	Target     TargetType `yaml:"target"`
}

// NewConfig sets up a chalog configuration with sensible defaults
func NewConfig() *Config {
	return &Config{
		In:         DefaultIn,
		Out:        DefaultOut,
		Repo:       DefaultRepo,
		Unreleased: DefaultUnreleased,
		Target:     DefaultTarget,
	}
}
