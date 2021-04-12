package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/jthomperoo/chalog/chalog"
	"github.com/jthomperoo/chalog/conf"
	"gopkg.in/yaml.v2"
)

// Version specifies the chalog tool version, overridden at build time
var Version string = "development"

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
	defaultConfig  = ".chalog.yml"
	defaultVersion = false
)

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

	inFlag := flag.String(flagIn, conf.DefaultIn,
		"the directory for storing the changelog files")
	outFlag := flag.String(flagOut, conf.DefaultOut,
		"the changelog file to output to")
	repoFlag := flag.String(flagRepo, conf.DefaultRepo,
		"the repository base url, include the protocol (http/https)")
	unreleasedFlag := flag.String(flagUnreleased, conf.DefaultUnreleased,
		"the release name that should be treated as a the 'unreleased' section")
	configFlag := flag.String(flagConfig, defaultConfig,
		"path to the config file to load")
	targetFlag := flag.String(flagTarget, string(conf.DefaultTarget),
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

	config := conf.NewConfig()
	err = yaml.Unmarshal([]byte(configData), config)
	if err != nil {
		log.Fatalln(err)
	}

	if isFlagPassed(flagIn) {
		config.In = *inFlag
	}
	if isFlagPassed(flagOut) {
		config.Out = *outFlag
	}
	if isFlagPassed(flagRepo) {
		config.Repo = *repoFlag
	}
	if isFlagPassed(flagUnreleased) {
		config.Unreleased = *unreleasedFlag
	}
	if isFlagPassed(flagTarget) {
		config.Target = conf.TargetType(*targetFlag)
	}

	generator := chalog.NewGenerator()

	output, err := generator.Generate(config)
	if err != nil {
		log.Fatalln(err)
	}

	switch config.Target {
	case conf.TargetTypeFile:
		err = ioutil.WriteFile(config.Out, []byte(output), 0644)
		if err != nil {
			log.Fatal(err)
		}
	case conf.TargetTypeStdout:
		fmt.Print(output)
	default:
		log.Fatalf("unknown target type '%s' provided", config.Target)
	}

}
