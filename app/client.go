package app

import (
	"bytes"
	"encoding/gob"
	// "log"
)

// Return a GOB decoder primed wiht the bytes for a given key and subkey 
func GetLoadedDecoder(key string, subkey string) (*gob.Decoder, error) {
	reply, err := GetBytes(key, subkey)

	if reply.Message == "critical" {
		return nil, err
	}

	p := bytes.NewBuffer(reply.Payload)
	//bytes.Buffer satisfies the interface for io.Writer and can be used
	//in gob.NewDecoder() 
	return gob.NewDecoder(p), nil
}

// Return an encoder and buffer ready to encode anything
func GetBufferEncoder() (*gob.Encoder, *bytes.Buffer) {
	// b := make([]byte, 0, 1024*1024*2)
	// m := bytes.NewBuffer(b)

	// log.Println(" Buffer Len: ", m.Len())

	m := new(bytes.Buffer)
	//the *bytes.Buffer satisfies the io.Writer interface and can
	//be used in gob.NewEncoder() 
	enc := gob.NewEncoder(m)

	return enc, m
}
