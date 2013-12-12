package app

import (
	"bytes"
	"encoding/gob"
	"strconv"
	"testing"
	"time"
)

type Xact struct {
	Description string
	Other       string
	Amount      int
	Date        int64
}

type XactList []Xact

type XactTime struct {
	Description string
	Other       string
	Amount      int
	Date        time.Time
}

type XactTimeList []XactTime

// alocated once to remove overhead
var xactEncoder *gob.Encoder
var xactBuffer *bytes.Buffer

func makeSomeXacts(incrementDate bool) *XactList {
	x1 := Xact{
		Description: "My very first description",
		Other:       "My other description",
		Amount:      1245,
		Date:        1232329394,
	}

	x2 := Xact{
		Description: "My very second description",
		Other:       "My other scond ",
		Amount:      12459,
		Date:        45343254326,
	}

	thing := make(XactList, 20000)

	for i := 0; i < 20000; i += 2 {
		x1.Description = "My Nicely seperated " + strconv.Itoa(i)
		if incrementDate {
			x1.Date = x1.Date + int64(i)
			x2.Date = x2.Date + int64(i+1)
		}
		thing[i] = x1
		thing[i+1] = x2
	}

	return &thing
}

func makeSomeTimeXacts() *XactTimeList {
	x1 := XactTime{
		Description: "My very first description",
		Other:       "My other description",
		Amount:      1245,
		Date:        time.Now(),
	}

	x2 := XactTime{
		Description: "My very second description",
		Other:       "My other scond ",
		Amount:      12459,
		Date:        time.Now(),
	}

	thing := make(XactTimeList, 20000)

	for i := 0; i < 20000; i += 2 {
		x1.Description = "My Nicely seperated " + strconv.Itoa(i)
		thing[i] = x1
		thing[i+1] = x2
	}

	return &thing
}

func PostPreAllocatedXactList(key string, subkey string, e *XactList) int {

	xactBuffer.Reset()
	xactEncoder.Encode(*e)
	data := xactBuffer.Bytes()

	return len(data)
}

func saveSomePreAllocatedXacts(thing *XactList) {

	PostPreAllocatedXactList("danmull", "fubar", thing)
	PostPreAllocatedXactList("ganmull", "fubar", thing)
	PostPreAllocatedXactList("ianmull", "fubar", thing)
	PostPreAllocatedXactList("kanmull", "fubar", thing)
	PostPreAllocatedXactList("lanmull", "fubar", thing)
	PostPreAllocatedXactList("nanmull", "fubar", thing)
	PostPreAllocatedXactList("panmull", "fubar", thing)
	PostPreAllocatedXactList("ranmull", "fubar", thing)
	PostPreAllocatedXactList("uanmull", "fubar", thing)
	PostPreAllocatedXactList("wanmull", "fubar", thing)
}

func Benchmark_EncodePreAllocated(b *testing.B) {

	// grab the encoder and buffer once
	xactEncoder, xactBuffer = GetBufferEncoder()

	thingy := makeSomeXacts(false)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		saveSomePreAllocatedXacts(thingy)

	}
}

// grab the buffer each time
func PostXactList(key string, subkey string, e *XactList) int {

	enc, m := GetBufferEncoder()
	enc.Encode(*e)

	data := m.Bytes()

	return len(data)
}

func saveSomeXacts(thing *XactList) {

	PostXactList("danmull", "fubar", thing)
	PostXactList("ganmull", "fubar", thing)
	PostXactList("ianmull", "fubar", thing)
	PostXactList("kanmull", "fubar", thing)
	PostXactList("lanmull", "fubar", thing)
	PostXactList("nanmull", "fubar", thing)
	PostXactList("panmull", "fubar", thing)
	PostXactList("ranmull", "fubar", thing)
	PostXactList("uanmull", "fubar", thing)
	PostXactList("wanmull", "fubar", thing)
}

func Benchmark_Encode(b *testing.B) {

	thingy := makeSomeXacts(false)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		saveSomeXacts(thingy)
	}
}

func Benchmark_Encode_Inc(b *testing.B) {

	thingy := makeSomeXacts(true)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		saveSomeXacts(thingy)
	}
}

func PostXactTimeList(key string, subkey string, e *XactTimeList) int {

	enc, m := GetBufferEncoder()
	enc.Encode(*e)

	data := m.Bytes()

	return len(data)
}

func saveSomeXactTimes(thing *XactTimeList) {

	PostXactTimeList("danmull", "fubar", thing)
	PostXactTimeList("ganmull", "fubar", thing)
	PostXactTimeList("ianmull", "fubar", thing)
	PostXactTimeList("kanmull", "fubar", thing)
	PostXactTimeList("lanmull", "fubar", thing)
	PostXactTimeList("nanmull", "fubar", thing)
	PostXactTimeList("panmull", "fubar", thing)
	PostXactTimeList("ranmull", "fubar", thing)
	PostXactTimeList("uanmull", "fubar", thing)
	PostXactTimeList("wanmull", "fubar", thing)
}

func Benchmark_Encode_With_Time(b *testing.B) {

	thingy := makeSomeTimeXacts()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		saveSomeXactTimes(thingy)
	}
}
