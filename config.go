package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"

	_ "embed"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var (
	debug    = flag.Bool("v", false, "debug mode")
	initMode = flag.Bool("init", false, "create init config file")
	//go:embed init.yml
	initConfigFile []byte

	defaultMaxErr uint    = 15
	defaultMaxRun         = 3
	defaultAlpha  float64 = 0.01

	runcount = 0
)

type Config struct {
	// Generated benchmark file (Required)
	File *string `yaml:"file"`
	// Execute this command before run
	BeforeRun *string `yaml:"beforerun"`
	// Run this command generate a new benchmark file (Required)
	Run *string `yaml:"run"`
	// Execute this command after run
	AfterRun *string `yaml:"afterrun"`
	// The maximum error for every benchmark
	Maxerr *uint `yaml:"maxerr"`
	// The maximum run count
	MaxRun *int `yaml:"maxrun"`
	// Ignored benchmark methods(comma-separated)
	Ignore *string `yaml:"ignore"`

	Alpha *float64 `yaml:"alpha"`
}

func (c *Config) SetDefault() {
	if c.File == nil {
		file := ""
		c.File = &file
	}
	if c.BeforeRun == nil {
		beforerun := ""
		c.BeforeRun = &beforerun
	}
	if c.Run == nil {
		run := ""
		c.Run = &run
	}
	if c.AfterRun == nil {
		afterrun := ""
		c.AfterRun = &afterrun
	}
	if c.Ignore == nil {
		ignore := ""
		c.Ignore = &ignore
	}
	if c.Maxerr == nil {
		maxerr := defaultMaxErr
		c.Maxerr = &maxerr
	}
	if c.MaxRun == nil {
		maxrun := defaultMaxRun
		c.MaxRun = &maxrun
	}
	if c.Alpha == nil {
		alpha := defaultAlpha
		c.Alpha = &alpha
	}
}

func (c Config) String() string {
	return fmt.Sprintf("\nfile: %s\nbeforerun: %s\nrun: %s\nafterrun: %s\nignore: %s\nmaxerr: %d\nmaxrun: %d\n", *c.File, *c.BeforeRun, *c.Run, *c.AfterRun, *c.Ignore, *c.Maxerr, *c.MaxRun)
}

func readConfig() {
	c := readLocalConfig(flag.Arg(0))
	if c != nil {
		// Use local config.
		globalConfig = *c
	}
	// Give default value if not set.
	globalConfig.SetDefault()

	// Print global config if debug.
	logrus.Debugln("result config:", globalConfig.String())

	validConfig()
}

func validConfig() {
	if *globalConfig.Run == "" {
		logrus.Fatalln("config: empty run")
	}
	if *globalConfig.File == "" {
		logrus.Fatalln("config: empty file")
	}
}

func readLocalConfig(name string) *Config {
	var (
		content []byte
		err     error
	)
	for _, v := range []string{".benchok.yml", ".benchok.yaml"} {
		content, err = os.ReadFile(v)
		if err == nil {
			break
		}
	}
	if content == nil || err != nil {
		logrus.Debugln("config:" + err.Error())
		return nil
	}

	var configs map[string]Config
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	decoder.KnownFields(true)
	err = decoder.Decode(&configs)
	if err != nil {
		logrus.Debugln("config:" + err.Error())
		return nil
	}

	for k, v := range configs {
		// Empty name use the first config.
		if k == name || name == "" {
			logrus.Infoln(fmt.Sprintf(`config: use config "%s"`, k))
			return &v
		}
	}

	if name != "" {
		logrus.Fatalln("config: can't find specified config", name)
		return nil
	}
	logrus.Debugln("config: no config")
	return nil
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
