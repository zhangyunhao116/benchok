package main

import (
	"errors"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
	"golang.org/x/perf/benchstat"
)

type Item struct {
	from       string  // FooQueue
	methodname string  // 50%Enqueue50%Dequeue
	cpu        int     // cpu numbers
	timeop     float64 // ns/op
	timeopstr  string  // 32.23 (default ".2f")
	delta      int     // %
	rawname    string
	rawline    string
}

func (i *Item) String() string {
	return i.rawname + " " + strconv.Itoa(i.delta) + "%"
}

func parseAllItem() (*benchstat.Collection, error) {
	c := &benchstat.Collection{
		Alpha:      *globalConfig.Alpha,
		AddGeoMean: false,
		DeltaTest: func(old, new *benchstat.Metrics) (float64, error) {
			return -1, nil
		},
	}
	f, err := os.Open(*globalConfig.File)
	if err != nil {
		return nil, err
	}
	if err := c.AddFile(*globalConfig.File, f); err != nil {
		return nil, err
	}
	f.Close()

	return c, nil
}

func parseAllItemOnlySpeed() (*benchstat.Table, error) {
	collection, err := parseAllItem()
	if err != nil {
		return nil, err
	}

	// Get table `time/op`.
	var speedTable *benchstat.Table
	for _, v := range collection.Tables() {
		if v.Metric == "time/op" {
			speedTable = v
			break
		}
	}

	if speedTable == nil {
		return nil, errors.New("can not find `time/op table`")
	}
	return speedTable, nil
}

// `5%` -> 5
func parseDiff(s string) int {
	res, err := strconv.Atoi(s[0 : len(s)-1])
	if err != nil {
		logrus.Fatalf("parseDiff: (%s) %s", s, err.Error())
	}
	return res
}
