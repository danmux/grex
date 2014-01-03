package app

// key index - the index of all current keys

import (
	"container/list"
	// "errors"
	"hash/fnv"
	"strings"
)

type KeyIndex struct {
	Maps [](map[string]int) // 0xFF maps
}

// func (me MappedKey) String() string {
// 	return fmt.Sprintf("%s-%010d", me.Key[0:3], me.Num)
// }

var keyIndex = KeyIndex{}

var HASHSIZE = 0x10000

func fuckup(v uint64) uint64 {
	// byte 7 constant        + these dont change and byte 0 -> byte 6 and    byte 1 -> byte 3   and          byte 7-> byte0               and byte 3-> byte1 
	return (0xF1 << (7 * 8)) + v&0x0000FFFF00FF0000 + ((v & 0xFF) << (6 * 8)) + ((v & 0xFF00) << 16) + ((v & 0x00FF000000000000) >> (6 * 8)) + ((v & 0xFF000000) >> 16)
}

func (me *KeyIndex) init() {
	// l := len(KEY_CHARS)
	// me.Keys = make([]*list.List, l*l*l)
	// for k0, _ := range KEY_CHARS {
	// 	for k1, _ := range KEY_CHARS {
	// 		for k2, _ := range KEY_CHARS {
	// 			me.Keys[k0*l*l+k1*l+k2] = list.New()

	// 		}
	// 	}
	// }

	// me.Keys = make([]*list.List, HASHSIZE)
	// for i := 0; i < HASHSIZE; i++ {
	// 	me.Keys[i] = list.New()
	// }
}

func toIndNum(char byte) int {
	return strings.IndexByte(KEY_CHARS, char)
}

func toIndex(key string) int {
	if len(key) < 3 {
		return -1
	}
	l := len(KEY_CHARS)

	k0 := toIndNum(key[0])
	k1 := toIndNum(key[1])
	k2 := toIndNum(key[2])

	if k0 < 0 || k1 < 0 || k2 < 0 {
		return -1
	}

	return k0*l*l + k1*l + k2
}

var hash = fnv.New32a()

func hashIndex(key string) uint32 {
	hash.Reset()

	_, error := hash.Write([]byte(key))
	if error != nil {
		println(error.Error())
		panic("Hash error")
	}

	done := hash.Sum32()

	return done & 0xFFFF
}

func findKey(key string, l *list.List) string {
	// for e := l.Front(); e != nil; e = e.Next() {
	// 	if e.Value.(*MappedKey).Key == key {
	// 		return e.Value.(*MappedKey).String()
	// 	}
	// }
	return ""
}

func printList(l *list.List) {
	// for e := l.Front(); e != nil; e = e.Next() {
	// 	print(e.Value.(*MappedKey).Key, " ")
	// }
	println("done")
}

// add the value to the index returning the 
func (me *KeyIndex) Add(key string) (string, error) {

	// i := toIndex(key)
	// i := hashIndex(key)
	// if i < 0 {
	// 	return "", errors.New("attempt to add a bad key")
	// }

	// l := me.Keys[i]
	// if findKey(key, l) != "" {
	// 	return "", errors.New("attempt to add a duplicate key")
	// }

	// mv := l.Front()
	// maxNum := 0

	// if mv != nil {
	// 	maxNum = mv.Value.(*MappedKey).Num
	// }

	// mk := MappedKey{
	// 	key,
	// 	maxNum + 1,
	// }

	// // so newest added at front of list
	// l.PushFront(&mk)

	// return mk.String(), nil

	return "", nil
}

func (me KeyIndex) Find(key string) (string, bool) {

	// // i := toIndex(key)
	// i := hashIndex(key)

	// if i < 0 {
	// 	return "", false
	// }

	// l := me.Keys[i]
	// mk := findKey(key, l)

	// if mk == "" {
	// 	return "", false
	// }

	// return mk, true
	return "", false
}

func (me KeyIndex) Print() {
	// l := len(KEY_CHARS)
	// // b := []byte{' ', ' ', ' '}

	// maxBucket := 0
	// mbKey := ""
	// mbI := 0

	// for k0, c0 := range KEY_CHARS {
	// 	for k1, c1 := range KEY_CHARS {
	// 		for k2, c2 := range KEY_CHARS {

	// 			s := string(c0) + string(c1) + string(c2)

	// 			print(s, "-")
	// 			ln := me.Keys[k0*l*l+k1*l+k2].Len()

	// 			if ln > maxBucket {
	// 				maxBucket = ln
	// 				mbKey = s
	// 				mbI = k0*l*l + k1*l + k2
	// 			}
	// 			println(ln)

	// 		}
	// 	}
	// }

	// for i := 0; i < HASHSIZE; i++ {
	// 	ln := me.Keys[i].Len()

	// 	if ln > maxBucket {
	// 		maxBucket = ln
	// 		mbKey = me.Keys[i].Front().Value.(*MappedKey).Key
	// 		mbI = i
	// 	}
	// 	// println(ln)
	// }

	// println("Max Bucket", mbKey, maxBucket)

	// li := me.Keys[mbI]
	// printList(li)

}

func (me KeyIndex) Print2() {
	// l := len(KEY_CHARS)
	// // b := []byte{' ', ' ', ' '}
	// for k0, c0 := range KEY_CHARS {
	// 	for k1, c1 := range KEY_CHARS {
	// 		tot := 0

	// 		for k2, c2 := range KEY_CHARS {

	// 			tot = tot + me.Keys[k0*l*l+k1*l+k2].Len()

	// 			if c2 == '_' {
	// 				s := string(c0) + string(c1)

	// 				print(s, " - ")
	// 				println(tot)
	// 			}
	// 		}
	// 	}
	// }
}
