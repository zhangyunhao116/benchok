package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

var (
	globalConfig Config
	globalResult *benchmarkResult
)

func main() {
	// Init config.
	flag.Parse()
	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	if *initMode { // init config file then return
		const configFileName = ".benchok.yml"
		if _, err := os.Stat(configFileName); errors.Is(err, os.ErrNotExist) {
			if err := ioutil.WriteFile(configFileName, initConfigFile, 0660); err != nil {
				logrus.Fatal("init config file error:", err)
			}
		} else {
			logrus.Fatal("init config exists: ", configFileName)
		}
		return
	}
	readConfig()
	// Start.
	execBeforeRun()
parse:
	logrus.Debugln("parseAllItems start")
	_, err := parseAllItemOnlySpeed()
	if err != nil {
		logrus.Debugln("parseAllItems: ", err.Error())
		execrun()
		goto parse
	}
	logrus.Debugln("parseAllItems success")
	var (
		exceedCount int
		exceedInfo  string
	)
	resultMerge()
	for name, diff := range globalResult.info {
		if diff > int(*globalConfig.Maxerr) && !matchIgnore(name) {
			if exceedCount == 0 {
				exceedInfo = fmt.Sprintln(name, fmt.Sprintf("%d%%", diff), "exceed", strconv.Itoa(int(*globalConfig.Maxerr))+"%")
			}
			exceedCount++
		}
	}
	if exceedCount != 0 {
		logrus.Infoln(fmt.Sprintf("(%d/%d)", exceedCount, len(globalResult.info)), exceedInfo)
		// Rerun.
		execrun()
		goto parse
	}
	logrus.Debugln("all success")
	resultWriteLocal()
	execAfterRun()
}

func execrun() {
	runcount++
	logrus.Debugf("run id: %d", runcount)
	if *globalConfig.MaxRun > 0 {
		if runcount > *globalConfig.MaxRun {
			resultWriteLocal()
			execAfterRun()
			logrus.Infoln("run: exceed maximum run count", *globalConfig.MaxRun)
			logrus.Exit(0)
		}
	}

	_, err := execCommandPrint("run", *globalConfig.Run)
	if err != nil {
		logrus.Fatalln(err.Error())
	}
}

func execBeforeRun() {
	if globalConfig.BeforeRun != nil && *globalConfig.BeforeRun != "" {
		logrus.Debugln("beforerun start")
		_, err := execCommandPrint("before", *globalConfig.BeforeRun)
		if err != nil {
			logrus.Fatalln("beforerun:", err.Error())
		}
		logrus.Debugln("beforerun success")
	}
}

func execAfterRun() {
	if globalConfig.AfterRun != nil && *globalConfig.AfterRun != "" {
		logrus.Debugln("afterrun start")
		_, err := execCommandPrint("afterrun", *globalConfig.AfterRun)
		if err != nil {
			logrus.Fatalln("afterrun:", err.Error())
		}
		logrus.Debugln("afterrun success")
	}
}

func resultMerge() {
	if globalResult == nil {
		res, err := newBenchmarkResult()
		if err != nil {
			logrus.Fatalln(err)
		}
		globalResult = res
	} else {
		err := globalResult.merge()
		if err != nil {
			logrus.Fatal(err)
		}
	}
}

func resultWriteLocal() {
	err := globalResult.writeLocal(*globalConfig.File)
	if err != nil {
		logrus.Fatal(err)
	}
}

func matchIgnore(s string) bool {
	if *globalConfig.Ignore == "" {
		return false
	}
	ignores := strings.Split(*globalConfig.Ignore, ",")
	for _, v := range ignores {
		if strings.Contains(s, v) {
			return true
		}
	}
	return false
}

func execCommand(prefix, cmd string) (string, error) {
	var stderr bytes.Buffer
	command := exec.Command("bash", "-c", cmd)
	command.Stderr = &stderr
	out, err := command.Output()
	if err != nil {
		return stderr.String(), errors.New(prefix + ": " + err.Error())
	}
	return string(out), nil
}

func execCommandPrint(prefix, cmd string) (string, error) {
	out, err := execCommand(prefix, cmd)
	if err != nil {
		logrus.Warningf(`exec "%s" error: `, prefix)
	}
	print(string(out))
	return out, err
}

func execCommandPrintOnlyFailed(prefix, cmd string) (string, error) {
	out, err := execCommand(prefix, cmd)
	if err != nil {
		logrus.Warningf(`exec "%s" error: `, prefix)
		print(string(out))
	}
	return out, err
}
