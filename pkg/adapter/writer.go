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

package adapter

import (
	"log"
	"strconv"

	"github.com/bragfoo/TiPrometheus/pkg/conf"
	"github.com/bragfoo/TiPrometheus/pkg/lib"

	"bytes"
	"time"

	"github.com/bragfoo/TiPrometheus/pkg/tikv"
	"github.com/prometheus/prometheus/prompb"
	"go.uber.org/zap/buffer"
)

var (
	buffers = buffer.NewPool()
)

func RemoteWriter(data prompb.WriteRequest) {
	for _, oneDoc := range data.Timeseries {
		labels := oneDoc.Labels
		samples := oneDoc.Samples
		log.Println("Naive write data:", labels, samples)

		//build index and return labelID
		labelID := buildIndex(labels, samples)
		//log.Println("LabelID:", labelID)

		//write timeseries data
		writeTimeseriesData(labelID, samples)

		// TODO: need handle error
		labelsByte, _ := lib.GetBytes(labels)
		SaveOriDoc(labelID, labelsByte)
	}
}

//build md5 data and store to kv if not exist
func buildIndex(labels []*prompb.Label, samples []prompb.Sample) string {
	//make md
	//key type key#value
	buf := buffers.Get()
	defer buf.Free()
	for _, v := range labels {
		buf.AppendString(v.Name)
		buf.AppendString("#")
		buf.AppendString(v.Value)
	}

	labelBytes := buf.Bytes()
	labelID := lib.MakeMDByByte(labelBytes)
	buf.Reset()

	//labels index
	for _, v := range labels {
		//key type index:label:__name__#latency
		buf.AppendString("index:label:")
		buf.AppendString(v.Name)
		buf.AppendString("#")
		buf.AppendString(v.Value)
		key := buf.String()
		buf.Reset()
		//log.Println("Write label md:", key, labelID)

		//key type index:status:__name__#latency+labelID
		buf.AppendString("index:status:")
		buf.AppendString(v.Name)
		buf.AppendString("#")
		buf.AppendString(v.Value)
		buf.AppendString("+")
		buf.AppendString(labelID)
		indexStatus := buf.Bytes()

		indexStatusKey, _ := tikv.Get([]byte(indexStatus))
		//log.Println("indexStatus:", indexStatusKey)

		//not in index
		if "" == indexStatusKey.Value {
			// TODO: need handle error
			err := tikv.Puts([]byte(indexStatus), []byte("1"))
			if err != nil {
				log.Print(err)
			}

			//wtire tikv
			// TODO: need handle error
			oldKey, err := tikv.Get([]byte(key))
			if err != nil {
				log.Print(err)
			}

			if oldKey.Value == "" {
				// TODO: need handle error
				err := tikv.Puts([]byte(key), []byte(labelID))
				if err != nil {
					log.Print(err)
				}
			} else {
				b := bytes.NewBufferString(oldKey.Value)
				b.WriteString(labelID)
				v := b.Bytes()

				// TODO: need handle error
				err := tikv.Puts([]byte(key), v)
				if err != nil {
					log.Print(err)
				}
			}
		}

		buf.Reset()
	}

	buf.Reset()
	buf.AppendString("index:timeseries:")

	now := time.Now().UnixNano() / int64(time.Millisecond)

	interval := int64(conf.RunTimeInfo.TimeInterval * 1000 * 60)
	now = (now / interval) * interval

	buf.AppendString(labelID)
	buf.AppendString(":")
	buf.AppendString(strconv.FormatInt(now, 10))

	timeIndexBytes := buf.Bytes()

	//timeseries index
	for _, v := range samples {
		oldKey, _ := tikv.Get(timeIndexBytes)
		//log.Println("Timeseries indexStatus:", oldKey)
		if oldKey.Value == "" {
			// TODO: need handle error
			err := tikv.Puts(timeIndexBytes, lib.Int64ToBytes(v.Timestamp))
			if err != nil {
				log.Print(err)
			}
		} else {
			bs := buffers.Get()
			bs.AppendString(oldKey.Value)
			bs.AppendString(strconv.FormatInt(v.Timestamp, 10))
			v := bs.Bytes()
			// TODO: need handle error
			err := tikv.Puts(timeIndexBytes, v)
			if err != nil {
				log.Print(err)
			}
			bs.Free()
		}
	}

	return labelID
}

func writeTimeseriesData(labelID string, samples []prompb.Sample) {
	buf := buffers.Get()
	defer buf.Free()
	for _, v := range samples {
		//key type timeseries:doc:labelMD#timestamp
		buf.AppendString("timeseries:doc:")
		buf.AppendString(labelID)
		buf.AppendString(":")
		buf.AppendString(strconv.FormatInt(v.Timestamp, 10))
		key := buf.Bytes()

		//write to tikv
		// TODO: need handle error
		err := tikv.Puts(key, []byte(strconv.FormatFloat(v.Value, 'E', -1, 64)))
		if err != nil {
			log.Print(err)
		}
		//log.Println("Write timeseries:", string(key), strconv.FormatFloat(v.Value, 'E', -1, 64))
		buf.Reset()
	}
}

func SaveOriDoc(labelID string, originalMsg []byte) {
	buf := buffers.Get()
	defer buf.Free()
	buf.AppendString("doc:")
	buf.AppendString(labelID)
	key := buf.Bytes()
	// TODO: need handle error
	err := tikv.Puts(key, originalMsg)
	if err != nil {
		log.Print(err)
	}
	//log.Println("Write meta:", string(key), string(originalMsg))
}
