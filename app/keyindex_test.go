package app

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"testing"
)

func IGNORE_Test_KeyIndex(t *testing.T) { //test function starts with "Test" and takes a pointer to type testing.T

	if toIndNum('0') != 0 {
		t.Errorf("toindNum not good. %v", toIndNum('0'))
	}

	if toIndNum('9') != 9 {
		t.Errorf("toindNum not good. %v", toIndNum('9'))
	}

	if toIndNum('a') != 10 {
		t.Errorf("toindNum not good. %v", toIndNum('a'))
	}

	if toIndNum('z') != 35 {
		t.Errorf("toindNum not good. %v", toIndNum('z'))
	}

	if toIndex("000") != 0 {
		t.Errorf("toIndex not good. %v", toIndex("000"))
	}

	if toIndex("00a") != 10 {
		t.Errorf("toIndex not good. %v", toIndex("00a"))
	}

	if toIndex("0a0qqqqqq") != 370 {
		t.Errorf("toIndex not good. %v", toIndex("0a0qqqqq"))
	}

	if toIndex("___") != (37*37*37)-1 {
		t.Errorf("toIndex not good. %v", toIndex("zzz"))
	}

	if toIndex("zZz") != -1 {
		t.Errorf("toIndex not good. %v", toIndex("zZz"))
	}
}

func IGNORE_Test_KeyHash(t *testing.T) { //test function starts with "Test" and takes a pointer to type testing.T

	if hashIndex("dan") != 12328 {
		t.Errorf("toIndex not good. %v", hashIndex("dan"))
	}

	if hashIndex("dani") != 63059 {
		t.Errorf("toIndex not good. %v", hashIndex("dani"))
	}
}

func IGNORE_Test_KeyHashAdd(t *testing.T) { //test function starts with "Test" and takes a pointer to type testing.T

	ki := KeyIndex{}
	ki.init()

	r, err := ki.Add("danm")
	if err != nil {
		t.Error("add error ", err)
	}

	if r != "dan-0000000001" {
		t.Error("add error ", r)
	}
}

func IGNORE_Test_Fuckup(t *testing.T) { //test function starts with "Test" and takes a pointer to type testing.T
	println(fuckup(0))
	println(fuckup(1))
	println(fuckup(2))
	println(fuckup(3))
	println(fuckup(4))
	println(fuckup(5))

	println(fuckup(6))

	if fuckup(0xFF) != 0xF1FF000000000000 {
		t.Error("fuckup failed")
	}

	//.. .. .. 0f 7e 4f 95 dd
	//F1 dd 00 0f 95 4f 7e 00

	if fuckup(66543654365) != 0xF1dd000f954f7e00 {
		t.Error("fuckup failed")
	}

	println(fuckup(66543654366))

}

func IGNORE_Test_KeyIndexAdd(t *testing.T) { //test function starts with "Test" and takes a pointer to type testing.T

	ki := KeyIndex{}
	ki.init()

	r, err := ki.Add("danm")
	if err != nil {
		t.Error("add error ", err)
	}

	if r != "dan-0000000001" {
		t.Error("add error ", r)
	}

	r, err = ki.Add("danmull")
	if err != nil {
		t.Error("add error ", err)
	}

	r, err = ki.Add("danmully")
	if err != nil {
		t.Error("add error ", err)
	}

	if r != "dan-0000000001" {
		t.Error("add error ", r)
	}

	r, err = ki.Add("danmully")

	if err == nil {
		t.Error("should have got an error adding duplicate key")
	}

	r, in := ki.Find("danmullx")
	if in {
		t.Error("Found a key that should not have been found")
	}

	v, in := ki.Find("danmull")
	if !in {
		t.Error("missing key danmull")
	}

	if v != "dan-0000000001" {
		t.Error("danmull has wrong key ", v)
	}

}

func IGNORE_Test_SnapChat(t *testing.T) {

	ki := KeyIndex{}
	ki.init()
	PrepRest()

	var err error
	file, err := os.Open("../testdata/schat/schat.csv")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file.Close()
	reader := csv.NewReader(file)

	cnt := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("Error:", err)
			return
		}

		// fmt.Println(record) // record has the type []string

		un := TidyKey(record[1])
		_, err = ki.Add(un)
		if err != nil {
			println(err.Error(), un, record[1])
		}
		cnt++

		if cnt%100000 == 0 {
			println("CNT --------------------------- ", cnt, (cnt*100)/4600000)
		}

		if cnt > 500000 {
			ki.Print()
			return
		}
	}

	ki.Print()

}

func IGNORE_Test_SnapChatNative(t *testing.T) {

	InitGrex("../testdata", "my.eg.uri", "8888", 10, 1)

	mp := make(map[string]int, 0x10000)
	// mp := make(map[string]int)
	PrepRest()

	var err error
	file, err := os.Open("../testdata/schat/schat.csv")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file.Close()
	reader := csv.NewReader(file)

	cnt := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("Error:", err)
			return
		}

		// fmt.Println(record) // record has the type []string

		un := TidyKey(record[1])

		_, in := mp[un]
		if !in {
			mp[un] = cnt
		} else {
			println("attempt to add duplicate key ", un)
		}

		cnt++

		if cnt%100000 == 0 {
			println("CNT --------------------------- ", cnt, (cnt*100)/4600000)
		}

		if cnt > 500000 {
			// ki.Print()
			break
		}
	}

	println(len(mp))

	PutObject("internal", "index", &mp)
	println(len(mp))

	md := make(map[string]int)
	GetObject("internal", "index", &md)

	println(len(md))

}
