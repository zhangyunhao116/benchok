package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
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

func parseAllItems() ([]*Item, error) {
	out, err := execCommand("benchstat: ", "benchstat -csv "+*globalConfig.File)
	if err != nil {
		if logrus.GetLevel() == logrus.DebugLevel {
			logrus.Warningln("parseAllItem stderr:")
			println(out)
		}
		return nil, err
	}
	inputStr := out
	// Parse csv.
	reader := csv.NewReader(bytes.NewReader([]byte(inputStr)))
	lines, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(lines) <= 1 {
		logrus.Fatalln(fmt.Sprintln("parse: too few lines: ", len(lines)))
	}
	firstLine := lines[0]
	if firstLine[0] != "name" || firstLine[1] != "time/op (ns/op)" || firstLine[2] != "±" {
		logrus.Fatalln(fmt.Sprintf("parse: invalid first line: %v, want (%s)", len(lines), "name,time/op (ns/op),±"))
	}

	var allItems []*Item
	// Parse all data.
	for _, v := range lines[1:] {
		// Default/70Enqueue30Dequeue/LinkedQ-100,1.00808E+02,3%
		if len(v) != 3 {
			logrus.Fatalln(fmt.Sprintf("parse: invalid line: %v", v))
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
			logrus.Fatalln(fmt.Sprintln("parse: invalid time/op: ", v[1]))
		}
		item.timeop = timeop
		item.timeopstr = fmt.Sprintf("%.2f", timeop)
		delta, err := strconv.Atoi(v[2][:len(v[2])-1])
		if err != nil {
			logrus.Fatalln(fmt.Sprintln("parse: invalid delta: ", v[2]))
		}
		item.delta = delta
		// Add this item.
		allItems = append(allItems, item)
	}
	return allItems, nil
}
