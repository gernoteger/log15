// Copyright 2016 Genot Eger+jmtuley.
// started off from https://gist.github.com/jmtuley/d4b09617967e59c58c3e, and modified by Gernot Eger

package gelf

import (
	"os"
	"strconv"
	"strings"

	"github.com/inconshreveable/log15"
)

// Handler sends logs to Graylog in GELF.
type GelfHandler struct {
	gelfWriter *Writer
	host       string
}


// like New but panics if error occurs
func MustNew(gelfAddr string) log15.Handler {
	h, err := NewGelfHandler(gelfAddr)
	if err != nil {
		panic(err)
	}
	return h
}

// Log forwards a log message to the specified receiever.
func (h GelfHandler) Log(r *log15.Record) error {
	short, full := shortAndFull(r.Msg)

	ctx := ctxToMap(r.Ctx)
	callerFile, callerLine := caller(ctx)
	delete(ctx, "_caller")

	m := &Message{
		Version:  "1.1",
		Host:     h.host,
		Short:    short,
		Full:     full,
		TimeUnix: float64(r.Time.UnixNano()/1000000) / 1000., // seconds with millis from record
		//TimeUnix: float64(r.Time.UnixNano())/1e9 ,		// full timestamp
		Level:    log15LevelsToSyslog[r.Lvl],
		File:     callerFile,
		Line:     callerLine,
		Extra:    ctx,
	}

	return h.gelfWriter.WriteMessage(m)
}

func ctxToMap(ctx []interface{}) map[string]interface{} {
	m := make(map[string]interface{}, len(ctx)/2)
	for i := 0; i < len(ctx); i += 2 {
		s := ctx[i].(string)
		m["_"+s] = ctx[i+1]
	}

	return m
}

func shortAndFull(msg string) (short string, full string) {
	lines := strings.SplitN(msg, "\n", 2)
	short = lines[0]
	if len(lines) > 1 {
		full = msg
	}

	return short, full
}

// caller searches a context list for an entry called "caller" and splits it into
// filename and line number.
func caller(ctx map[string]interface{}) (string, int) {
	info, present := ctx["_caller"]
	if !present {
		return "", 0
	}

	parts := strings.Split(info.(string), ":")
	line, _ := strconv.Atoi(parts[1])
	return parts[0], line
}

// source: http://www.cisco.com/c/en/us/td/docs/security/asa/syslog-guide/syslogs/logsevp.html
var log15LevelsToSyslog = map[log15.Lvl]int32{
	log15.LvlCrit:  2,
	log15.LvlError: 3,
	log15.LvlWarn:  4,
	log15.LvlInfo:  6,
	log15.LvlDebug: 7,
}
