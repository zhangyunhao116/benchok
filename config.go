package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var (
	flags = []string{"file", "maxerr", "maxrun", "ignore", "run"}

	debug     = flag.Bool("v", false, "debug mode")
	file      = flag.String("file", "", "Generated benchmark file (Required)")
	maxerr    = flag.Uint("maxerr", defaultMaxErr, "The maximum error for every benchmark (Required)")
	maxrun    = flag.Int("maxrun", defaultMaxRun, "The maximum run count")
	ignore    = flag.String("ignore", "", "Ignored benchmark methods(comma-separated)")
	beforerun = flag.String("beforerun", "", "Execute this command before run")
	run       = flag.String("run", "", "Run this command generate a new benchmark file (Required)")
	afterrun  = flag.String("afterrun", "", "Execute this command after run")

	defaultMaxErr uint = 15
	defaultMaxRun      = -1

	runcount = 0
)

type Config struct {
	File      *string `yaml:"file"`
	BeforeRun *string `yaml:"beforerun"`
	Run       *string `yaml:"run"`
	AfterRun  *string `yaml:"afterrun"`
	Maxerr    *uint   `yaml:"maxerr"`
	MaxRun    *int    `yaml:"maxrun"`
	Ignore    *string `yaml:"ignore"`
}

func (c Config) String() string {
	var (
		file, beforerun, run, afterrun, ignore string
		maxerr                                 uint
		maxrun                                 int
	)
	if c.File != nil {
		file = *c.File
	}
	if c.BeforeRun != nil {
		beforerun = *c.BeforeRun
	}
	if c.Run != nil {
		run = *c.Run
	}
	if c.AfterRun != nil {
		afterrun = *c.AfterRun
	}
	if c.Ignore != nil {
		ignore = *c.Ignore
	}
	if c.Maxerr != nil {
		maxerr = *c.Maxerr
	}
	if c.MaxRun != nil {
		maxrun = *c.MaxRun
	}
	return fmt.Sprintf("\nfile: %s\nbeforerun: %s\nrun: %s\nafterrun: %s\nignore: %s\nmaxerr: %d\nmaxrun: %d\n", file, beforerun, run, afterrun, ignore, maxerr, maxrun)
}

func readConfig() {
	c := readLocalConfig(flag.Arg(0))
	if c != nil {
		// Use local config.
		globalConfig.File = c.File
		globalConfig.MaxRun = c.MaxRun
		globalConfig.Maxerr = c.Maxerr
		globalConfig.Ignore = c.Ignore
		globalConfig.BeforeRun = c.BeforeRun
		globalConfig.Run = c.Run
		globalConfig.AfterRun = c.AfterRun
		// Give default value if not set.
		if globalConfig.Maxerr == nil {
			globalConfig.Maxerr = &defaultMaxErr
		}
		if globalConfig.MaxRun == nil {
			globalConfig.MaxRun = &defaultMaxRun
		}
	}
	// Use command-line config if possible.
	if isFlagPassed("file") {
		logrus.Debug("use command-line -file")
		globalConfig.File = file
	}
	if isFlagPassed("beforerun") {
		logrus.Debug("use command-line -beforerun")
		globalConfig.BeforeRun = beforerun
	}
	if isFlagPassed("run") {
		logrus.Debug("use command-line -run")
		globalConfig.Run = run
	}
	if isFlagPassed("afterrun") {
		logrus.Debug("use command-line -afterrun")
		globalConfig.AfterRun = afterrun
	}
	if isFlagPassed("maxrun") {
		logrus.Debug("use command-line -maxrun")
		globalConfig.MaxRun = maxrun
	}
	if isFlagPassed("maxerr") {
		logrus.Debug("use command-line -maxerr")
		globalConfig.Maxerr = maxerr
	}
	if isFlagPassed("ignore") {
		logrus.Debug("use command-line -ignore")
		globalConfig.Ignore = ignore
	}
	// Print global config if debug.
	logrus.Debugln("result config:", globalConfig.String())

	validConfig()
}

func validConfig() {
	if globalConfig.Run == nil || *globalConfig.Run == "" {
		logrus.Fatalln("config: empty run")
	}
	if globalConfig.File == nil || *globalConfig.File == "" {
		logrus.Fatalln("config: empty file")
	}
}

func readLocalConfig(name string) *Config {
	content, err := os.ReadFile(".benchok.yml")
	if err != nil {
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
			logrus.Infoln(fmt.Sprintf(`config: use local config "%s"`, k))
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
