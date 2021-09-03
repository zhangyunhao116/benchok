package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

type benchmarkResult struct {
	rawrows       []string
	benchmarkInfo map[string]int
}

func (r *benchmarkResult) writeLocal(s string) error {
	f, err := os.OpenFile(s, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	err = f.Truncate(0)
	if err != nil {
		return err
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		return err
	}
	_, err = f.Write([]byte(strings.Join(r.rawrows, "\n")))
	if err != nil {
		return err
	}
	return nil
}

func (r *benchmarkResult) merge() error {
	logrus.Debugln("merge begin")
	res, err := newBenchmarkResult()
	if err != nil {
		return err
	}

	for newname, newdiff := range res.benchmarkInfo {
		if r.benchmarkInfo[newname] > newdiff {
			// Do replace.
			logrus.Debugln("merge:", newname, r.benchmarkInfo[newname], "to", newdiff)
			r.benchmarkInfo[newname] = newdiff
			for i, lines := range r.rawrows {
				// Assert len(new) == len(old)
				if strings.Index(lines+" ", newname) != -1 {
					println("!!!")
					r.rawrows[i] = res.rawrows[i]
				}
			}
		}
	}
	r.writeLocal(fmt.Sprintf("%d.txt", runcount))
	return nil
}

func newBenchmarkResult() (*benchmarkResult, error) {
	result, err := os.ReadFile(*globalConfig.File)
	if err != nil {
		return nil, err
	}
	r := new(benchmarkResult)
	r.benchmarkInfo = make(map[string]int, 100)
	r.rawrows = strings.Split(string(result), "\n")

	table, err := parseAllItemOnlySpeed()
	if err != nil {
		return nil, err
	}

	for _, row := range table.Rows {
		r.benchmarkInfo[row.Benchmark] = parseDiff(row.Metrics[0].FormatDiff())
	}

	return r, nil
}
