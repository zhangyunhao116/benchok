package main

import (
	"bytes"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var (
	maxerr = flag.Int("maxerr", 15, "The maximum error for every benchmark")
	maxrun = flag.Int("maxrun", -1, "The maximum run count")
	ignore = flag.String("ignore", "", "Ignored benchmark methods")
	run    = flag.String("run", "", "Run this command if exceed maximum error")

	runcount = 0
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

func main() {
	flag.Parse()
	if flag.Arg(0) == "" {
		panic("need bench file")
	}
begin:
	allItems, err := parseAllItems()
	if err != nil {
		println("parseAllItems error:", err.Error())
		err = rerun()
		if err != nil {
			panic(err)
		}
		goto begin
	}
	for _, i := range allItems {
		if i.delta > *maxerr && !matchIgnore(i.rawname) {
			println(i.String(), "exceed", strconv.Itoa(*maxerr)+"%")
			if *run == "" {
				os.Exit(1)
				return
			}
			// Rerun.
			err = rerun()
			if err != nil {
				panic(err)
			}
			goto begin
		}
	}
}

func parseAllItems() ([]*Item, error) {
	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)
	cmd := exec.Command("bash", "-c", "benchstat -csv "+flag.Arg(0))
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	inputStr := stdout.String()
	// Parse csv.
	reader := csv.NewReader(bytes.NewReader([]byte(inputStr)))
	lines, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(lines) <= 1 {
		panic(fmt.Sprintln("too few lines: ", len(lines)))
	}
	firstLine := lines[0]
	if firstLine[0] != "name" || firstLine[1] != "time/op (ns/op)" || firstLine[2] != "±" {
		panic(fmt.Sprintf("invalid first line: %v, want (%s)", len(lines), "name,time/op (ns/op),±"))
	}

	var allItems []*Item
	// Parse all data.
	for _, v := range lines[1:] {
		// Default/70Enqueue30Dequeue/LinkedQ-100,1.00808E+02,3%
		if len(v) != 3 {
			panic(fmt.Sprintf("invalid line: %v", v))
		}
		item := new(Item)
		name := v[0]
		item.rawname = name
		item.rawline = strings.Join(v, ",")
		// Find CPU numbers.
		nameFindSnake := strings.LastIndex(name, "-")
		if nameFindSnake != -1 {
			cpu, err := strconv.Atoi(name[nameFindSnake+1:])
			if err == nil {
				item.cpu = cpu
			}
		} else {
			item.cpu = 1
		}
		if nameFindSnake != -1 && item.cpu != 1 { // remove "-128", 128 is the CPU numbers
			name = name[:nameFindSnake]
		}
		// Find from.
		nameFindFrom := strings.LastIndex(name, "/")
		if nameFindFrom == -1 {
			item.from = name
			item.methodname = name
		} else {
			item.from = name[nameFindFrom+1:]
			item.methodname = name[:nameFindFrom]
		}
		// Find timeop.
		timeop, err := strconv.ParseFloat(v[1], 64)
		if err != nil {
			panic(fmt.Sprintln("invalid time/op: ", v[1]))
		}
		item.timeop = timeop
		item.timeopstr = fmt.Sprintf("%.2f", timeop)
		delta, err := strconv.Atoi(v[2][:len(v[2])-1])
		if err != nil {
			panic(fmt.Sprintln("invalid delta: ", v[2]))
		}
		item.delta = delta
		// Add this item.
		allItems = append(allItems, item)
	}
	return allItems, nil
}

func rerun() error {
	runcount++
	if *maxrun > 0 {
		if runcount >= *maxrun {
			return errors.New("exceed maximum run count")
		}
	}
	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)
	cmd := exec.Command("bash", "-c", *run)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func matchIgnore(s string) bool {
	if *ignore == "" {
		return false
	}
	ignores := strings.Split(*ignore, ",")
	for _, v := range ignores {
		if strings.Contains(s, v) {
			return true
		}
	}
	return false
}
