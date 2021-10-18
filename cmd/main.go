package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

type SpanContext struct {
	TraceID string
	SpanID  string
}
type TraceEvent struct {
	Name         string
	SpanContext  SpanContext
	ParentSpanID string
	StartTime    time.Time
	EndTime      time.Time
	// don't really care about everything else
}

func usage() {
	fmt.Printf("usage: %s FILE\n", os.Args[0])
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	tracefilename := os.Args[1]
	tracedata, err := os.ReadFile(tracefilename)
	if err != nil {
		panic("unable to read trace file (" + tracefilename + "): " + err.Error())
	}
	var events []TraceEvent
	//	var event TraceEvent
	for {
		var event TraceEvent
		err = json.Unmarshal(tracedata, &event)
		if err == nil {
			events = append(events, event)
			break
		} else if jerr, ok := err.(*json.SyntaxError); ok && jerr.Offset > 0 && strings.Contains(jerr.Error(), "after top-level value") {
			json.Unmarshal(tracedata[:jerr.Offset-1], &event)
			events = append(events, event)
			tracedata = tracedata[jerr.Offset-1:]
			continue
		} else {
			panic("unable to marshal trace events: " + err.Error())
		}
	}
	eventMap := make(map[string]map[string]*TraceEvent)
	for i := range events {
		evt := &events[i]
		traceid := evt.SpanContext.TraceID
		spanid := evt.SpanContext.SpanID
		if eventMap[traceid] == nil {
			eventMap[traceid] = make(map[string]*TraceEvent)
		}
		eventMap[traceid][spanid] = evt
	}
	for _, spans := range eventMap {
		for _, span := range spans {
			id := span.Name
			dur := span.EndTime.Sub(span.StartTime)
			for span != nil && span.ParentSpanID != "" && span.ParentSpanID != "0000000000000000" {
				span = spans[span.ParentSpanID]
				if span != nil {
					id = span.Name + ";" + id
				}
			}
			fmt.Printf("%s %d\n", id, dur)
		}
	}
}
