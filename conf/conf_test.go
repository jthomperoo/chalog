package conf_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/chalog/conf"
)

func TestNewConfig(t *testing.T) {
	var tests = []struct {
		description string
		expected    *conf.Config
	}{
		{
			description: "Load defaults",
			expected: &conf.Config{
				In:         conf.DefaultIn,
				Out:        conf.DefaultOut,
				Repo:       conf.DefaultRepo,
				Unreleased: conf.DefaultUnreleased,
				Target:     conf.DefaultTarget,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result := conf.NewConfig()
			if !cmp.Equal(result, test.expected) {
				t.Errorf("result mismatch (-want +got):\n%s", cmp.Diff(test.expected, result))
			}
		})
	}
}
