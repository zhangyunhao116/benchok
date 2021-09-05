package main

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/zhangyunhao116/skipset"
)

type benchmarkResult struct {
	prefix   []string
	suffix   []string
	roworder []string            // row's order
	rows     map[string][]string // "Uint64/50Dequeue50Enqueue/LSCQ-8" : rows
	info     map[string]int      // "Uint64/50Dequeue50Enqueue/LSCQ-8"  : 5 (%)
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

	// Add prefix.
	var content string
	content += strings.Join(r.prefix, "\n")

	// Add rows.
	var rows []string
	for _, benchmethod := range r.roworder {
		rows = append(rows, r.rows[benchmethod]...)
	}
	content += "\n" + strings.Join(rows, "\n")

	// Add suffix.
	content += "\n" + strings.Join(r.suffix, "\n")

	_, err = f.Write([]byte(content))
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

	for benchmarkMethodName, diff := range res.info {
		if r.info[benchmarkMethodName] > int(*globalConfig.Maxerr) && r.info[benchmarkMethodName] > diff {
			// Do replace.
			logrus.Debugln("merge:", benchmarkMethodName, r.info[benchmarkMethodName], "to", diff)
			r.info[benchmarkMethodName] = diff
			r.rows[benchmarkMethodName] = res.rows[benchmarkMethodName]
		}
	}
	return nil
}

func newBenchmarkResult() (*benchmarkResult, error) {
	result, err := os.ReadFile(*globalConfig.File)
	if err != nil {
		return nil, err
	}

	r := new(benchmarkResult)
	// Init fields.
	r.info = make(map[string]int)
	r.prefix = make([]string, 0)
	r.suffix = make([]string, 0)
	r.roworder = make([]string, 0)
	r.rows = make(map[string][]string)

	// Init info.
	allBenchmark := skipset.NewString()
	table, err := parseAllItemOnlySpeed()
	if err != nil {
		return nil, err
	}
	for _, row := range table.Rows {
		r.info[row.Benchmark] = parseDiff(row.Metrics[0].FormatDiff())
		allBenchmark.Add(row.Benchmark)
	}

	// Init prefix.
	allrows := strings.Split(string(result), "\n")
	for _, row := range allrows {
		for _, v := range []string{"goos: ", "goarch: ", "pkg: ", "cpu: "} {
			if strings.Index(row, v) == 0 {
				r.prefix = append(r.prefix, row)
			}
		}
		for _, v := range []string{"PASS", "ok "} {
			if strings.Index(row, v) == 0 {
				r.suffix = append(r.suffix, row)
			}
		}
	}

	// Init row and roworder.
	for _, row := range allrows {
		allBenchmark.Range(func(benchmarkMethod string) bool {
			pre := "Benchmark" + benchmarkMethod + " "
			if strings.Index(row, pre) == 0 {
				// Add this row.
				if r.rows[benchmarkMethod] == nil {
					// First line.
					r.roworder = append(r.roworder, benchmarkMethod)
				}
				r.rows[benchmarkMethod] = append(r.rows[benchmarkMethod], row)
				return false
			}
			return true
		})
	}

	return r, nil
}
