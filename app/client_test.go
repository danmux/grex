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

func (t Xact) Size() int {
	return 50
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

func Test_PutObj(t *testing.T) {
	// make sure the cache isnt allocated
	seshCache = nil

	InitGrex("../testdata", "my.eg.uri", "8888", 10, 0)

	w := Xact{
		"xact 1",
		"something",
		1103,
		123254326245,
	}

	msg, err := PutObject("danm", "testobj", w)

	if err != nil {
		t.Error("error from PutObject", err)
	}
	if msg != "good" {
		t.Error("not got good response from PutObject", msg)
	}

	xx := Xact{}
	err = GetObject("danm", "testobj", &xx)
	if err != nil {
		t.Error("not got good response from GetObject", err)
	}
}

func Test_PutCachedObj(t *testing.T) {
	// make sure the cache isnt allocated
	seshCache = nil

	InitGrex("../testdata", "my.eg.uri", "8888", 10, 0)

	w := Xact{
		"xact 2",
		"something new",
		1135,
		1000054326245,
	}

	msg, err := PutCachedObject("fanm", "testobj", &w)

	if err != nil {
		t.Error("error from PutCachedObject", err)
	}
	if msg != "good" {
		t.Error("not got good response from PutCachedObject", msg)
	}

	y, in := GetItemFromCache("fanm", "testobj")

	if !in {
		t.Error("cache missing key fanm testobj")
	}

	if y.(*Xact).Description != "xact 2" {
		t.Error("cache returned some dodgy stuff")
	}

	xx := Xact{}
	newx, err := GetCachedObject("fanm", "testobj", &xx)
	if err != nil {
		t.Error("not got good response from GetObject", err)
	}

	if newx.(*Xact).Description != "xact 2" {
		t.Error("cache returned some dodgy stuff")
	}

	// remove object from the cache
	invalidateItemInCache("fanm", "testobj")

	z, inagain := GetItemFromCache("fanm", "testobj")

	if inagain {
		t.Error("obj not removed from cache")
	}

	if z != nil {
		t.Error("obj not nil so not removed from cache")
	}

	yx := Xact{}

	err = GetObject("fanm", "testobj", &yx)
	if err != nil {
		t.Error("not got good response from GetObject", err)
	}

	if yx.Description != "xact 2" {
		t.Error("cache returned some dodgy stuff")
	}

	yy := Xact{}
	newy, err := GetCachedObject("fanm", "testobj", &yy)
	if err != nil {
		t.Error("not got good response from GetCachedObject", err)
	}

	if newy.(*Xact).Description != "xact 2" {
		t.Error("cache returned some dodgy stuff")
	}

	newz, err := GetCachedObject("fanm", "testobj", &yy)
	if err != nil {
		t.Error("not got good response from GetCachedObject", err)
	}

	if newz.(*Xact).Description != "xact 2" {
		t.Error("cache returned some dodgy stuff")
	}

	// msg, err = PutCachedObject("fanm", "testobj", w)

}

func Test_GetCachedMissingObj(t *testing.T) {
	// make sure the cache isnt allocated
	seshCache = nil

	InitGrex("../testdata", "my.eg.uri", "8888", 10, 0)

	yy := Xact{}
	newy, err := GetCachedObject("xxanm", "testobj", &yy)
	if err == nil {
		t.Error("Should have gor an error")
	}

	if newy.(*Xact).Description != "" {
		t.Error("cache returned some dodgy stuff")
	}
}

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
