package lib

import (
	"bytes"
	"crypto/md5"
	"encoding/gob"
	"encoding/hex"
	"strconv"
)

func GetBytes(key interface{}) []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(key)
	return buf.Bytes()
}

func Int64ToBytes(i int64) []byte {
	return []byte(strconv.FormatInt(i, 10))
}

func MakeMDByByte(initByte []byte) string {
	m := md5.New()
	m.Write(initByte)
	md := m.Sum(nil)
	mdString := hex.EncodeToString(md)
	return mdString
}
