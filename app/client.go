package app

import (
	"bytes"
	"encoding/gob"
)

// Return a GOB decoder primed wiht the bytes for a given key and subkey 
func GetDecoder(key string, subkey string) (*gob.Decoder, error) {
	reply, err := GetBytes(key, subkey)

	if err != nil {
		return nil, err
	}

	p := bytes.NewBuffer(reply.Payload)
	//bytes.Buffer satisfies the interface for io.Writer and can be used
	//in gob.NewDecoder() 
	return gob.NewDecoder(p), nil
}

// Return an encoder and buffer ready to encode anything
func GetBufferEncoder() (*gob.Encoder, *bytes.Buffer) {
	m := new(bytes.Buffer)
	//the *bytes.Buffer satisfies the io.Writer interface and can
	//be used in gob.NewEncoder() 
	enc := gob.NewEncoder(m)

	return enc, m
}
