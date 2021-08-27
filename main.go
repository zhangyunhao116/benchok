package main

import (
	"bytes"
	"errors"
	"flag"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

var globalConfig Config

func main() {
	flag.Parse()
	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	readConfig()
begin:
	allItems, err := parseAllItems()
	if err != nil {
		logrus.Warningln("parseAllItems: ", err.Error())
		err = rerun()
		if err != nil {
			logrus.Fatalln(err)
		}
		goto begin
	}
	for _, i := range allItems {
		if i.delta > int(*globalConfig.Maxerr) && !matchIgnore(i.rawname) {
			logrus.Info(i.String(), "exceed", strconv.Itoa(int(*globalConfig.Maxerr))+"%")
			if *globalConfig.Run == "" {
				os.Exit(1)
				return
			}
			// Rerun.
			err = rerun()
			if err != nil {
				logrus.Fatalln(err)
			}
			goto begin
		}
	}
}

func rerun() error {
	logrus.Debugln("rerun")
	runcount++
	if *globalConfig.MaxRun > 0 {
		if runcount >= *globalConfig.MaxRun {
			return errors.New("exceed maximum run count")
		}
	}
	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)
	cmd := exec.Command("bash", "-c", *globalConfig.Run)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func matchIgnore(s string) bool {
	if globalConfig.Ignore == nil || *globalConfig.Ignore == "" {
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
