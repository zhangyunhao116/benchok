package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
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
	readConfig()
	// Start.
	execBeforeRun()
parse:
	logrus.Debugln("parseAllItems start")
	_, err := parseAllItemOnlySpeed()
	if err != nil {
		logrus.Debugln("parseAllItems: ", err.Error())
		err = execrun()
		if err != nil {
			logrus.Fatalln(err)
			return
		}
		goto parse
	}
	logrus.Debugln("parseAllItems success")
	if globalResult == nil {
		res, err := newBenchmarkResult()
		if err != nil {
			logrus.Fatalln(err)
		}
		globalResult = res
	} else {
		globalResult.merge()
	}
	var (
		exceedCount int
		exceedInfo  string
	)
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
		err = execrun()
		if err != nil {
			logrus.Fatalln(err)
			return
		}
		goto parse
	}
	logrus.Debugln("all success")
	err = globalResult.merge()
	if err != nil {
		logrus.Fatal(err)
		return
	}
	err = globalResult.writeLocal(*globalConfig.File)
	if err != nil {
		logrus.Fatal(err)
		return
	}
	execAfterRun()
}

func execrun() error {
	runcount++
	logrus.Debugf("run id: %d", runcount)
	if *globalConfig.MaxRun > 0 {
		if runcount >= *globalConfig.MaxRun {
			return errors.New("exceed maximum run count")
		}
	}

	_, err := execCommandPrint("run", *globalConfig.Run)
	if err != nil {
		return err
	}
	return nil
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
