package gelf

import (
	"github.com/go-stack/stack"
	"github.com/inconshreveable/log15"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"fmt"
	"github.com/stretchr/testify/require"
)

const SyslogInfoLevel = 6
const SyslogErrorLevel = 7

func TestWritingSimpleMessageToLocalUDP(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	r, err := NewReader("127.0.0.1:0")
	if err != nil {
		t.Fatalf("NewReader: %s", err)
	}

	loc, err := time.LoadLocation("Europe/Vienna")
	require.Nil(err)
	logTime := time.Now()
	logTime = time.Date(2016, 11, 23, 13, 01, 02, 123100*1e3, loc)

	msgData := "test message\nsecond line"
	rec := log15.Record{
		Time: logTime, //TODO: set fixed!!
		Lvl:  log15.LvlInfo,
		Msg:  msgData,
		Ctx:  []interface{}{"foo", "bar", "withField", "1"}, // no fields yet
		Call: stack.Caller(0),
	}

	h := MustNew(r.Addr())

	h.Log(&rec)

	msg, err := r.ReadMessage()

	require.Nil(err, "ReadMessage")

	assert.Equal("test message", msg.Short, "ReadMessage")
	assert.Equal(msgData, msg.Full, "ReadMessage")
	assert.EqualValues(SyslogInfoLevel, msg.Level, "ReadMessage")

	assert.EqualValues(2, len(msg.Extra), "number of extra fields")
	assert.EqualValues("handler_test.go", msg.File, "msg.File")
	//assert.EqualValues(32, msg.Line, "msg.Line") // quite instable, since it depends on line in code...

	extra := map[string]string{"foo": "bar", "withField": "1"}

	for k, v := range extra {
		// extra fields are prefixed with "_"
		val, ok := msg.Extra["_"+k].(string)
		assert.True(ok, "extra key exists: "+k)
		if ok {
			assert.EqualValues(v, val, "extra message "+k)
		}
	}

	// checking time...
	s := int64(msg.TimeUnix)
	ns := int64((msg.TimeUnix - float64(s)) * 1e9)
	mt := time.Unix(s, ns)

	//fmt.Printf("t0=%v time=%v t=%v", msg.TimeUnix, mt, logTime); fmt.Println()
	diff := logTime.Sub(mt)
	fmt.Printf("diff=%v", diff)
	fmt.Println()
	assert.WithinDuration(logTime, mt, time.Millisecond, "time from log") // we have millisecond precision

}
