// +build integration

package integration_test

import (
	"bytes"
	"io/ioutil"
	"log"
	"os/exec"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestIntegration(t *testing.T) {
	cmd := exec.Command("go", "run", "../main.go", "-target", "stdout", "-config", ".test_config.yml")

	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()

	if err != nil {
		log.Println(err.Error())
		t.Errorf("%v: %s", err, stderr.String())
		return
	}

	result := out.String()

	expectedBytes, err := ioutil.ReadFile("EXPECTED.MD")
	if err != nil {
		t.Error(err)
	}

	if !cmp.Equal(string(expectedBytes), result) {
		t.Errorf("metrics mismatch (-want +got):\n%s", cmp.Diff(string(expectedBytes), result))
	}
}
