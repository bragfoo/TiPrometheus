// Copyright 2021 The TiPrometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lib

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/gob"
	"encoding/hex"
	"strconv"

	"go.uber.org/zap/buffer"
)

var (
	buffers = buffer.NewPool()
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

func Int64WriteBytes(i int64) []byte {
	buf := buffers.Get()
	defer buf.Free()
	binary.Write(buf, binary.BigEndian, i)
	b := buf.Bytes()
	return b
}

func MakeMDByByte(initByte []byte) string {
	m := md5.New()
	m.Write(initByte)
	md := m.Sum(nil)
	mdString := hex.EncodeToString(md)
	return mdString
}

func ReadStringByStepwidth(step int, str string) []string {
	var buf []string
	for i := 0; i < len(str); i += step {
		buf = append(buf, str[i:i+step])
	}
	return buf
}

func ReadFixedLength(step int, bts []byte) []string {
	var buf []string
	for i := 0; i < len(bts); i += step {
		buf = append(buf, string(bts[i:i+step]))
	}
	return buf
}
