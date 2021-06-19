package main

import (
	"bytes"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
)

func transformEncoding(rawReader io.Reader, t transform.Transformer) (string, error) {
	ret, err := ioutil.ReadAll(transform.NewReader(rawReader, t))
	if err == nil {
		return string(ret), nil
	} else {
		return "Unknown character type", err
	}
}

// BytesFromShiftJIS converts an array of bytes (a valid ShiftJIS string) to a UTF-8 string
func BytesFromShiftJIS(b []byte) (string, error) {
	return transformEncoding(bytes.NewReader(b), japanese.ShiftJIS.NewDecoder())
}
