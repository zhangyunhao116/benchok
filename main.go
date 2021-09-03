package main

import (
	"bytes"
	"errors"
	"flag"
	"os/exec"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

var globalConfig Config

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
	speedTable, err := parseAllItemOnlySpeed()
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
	for _, row := range speedTable.Rows {
		diff := parseDiff(row.Metrics[0].FormatDiff())
		if diff > int(*globalConfig.Maxerr) && !matchIgnore(row.Benchmark) {
			logrus.Infoln(row.Benchmark, row.Metrics[0].FormatDiff(), "exceed", strconv.Itoa(int(*globalConfig.Maxerr))+"%")
			// Rerun.
			err = execrun()
			if err != nil {
				logrus.Fatalln(err)
				return
			}
			goto parse
		}
	}
	logrus.Debugln("all success")
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
