package adapter

import (
	"bytes"
	"github.com/bragfoo/TiPrometheus/src/lib"
	"github.com/bragfoo/TiPrometheus/src/modules/prompb"
	"github.com/bragfoo/TiPrometheus/src/modules/tikv"
	"log"
	"strconv"
	"time"
)

func RemoteWriter(data prompb.WriteRequest) {
	for _, oneDoc := range data.Timeseries {
		labels := oneDoc.Labels
		samples := oneDoc.Samples
		log.Println("Naive write data:", labels, samples)

		//build index and return labelID
		labelID := buildIndex(labels, samples)
		log.Println("LabelID:", labelID)

		//write timeseries data
		writeTimeseriesData(labelID, samples)

		labelsByte := lib.GetBytes(labels)
		SaveOriDoc(labelID, labelsByte)
	}
}

//build md5 data and store to kv if not exist
func buildIndex(labels []*prompb.Label, samples []*prompb.Sample) string {
	//make md
	var buffer = bytes.NewBufferString("")
	//key type key#value
	for _, v := range labels {
		buffer.WriteString(v.Name)
		buffer.WriteString("#")
		buffer.WriteString(v.Value)
	}
	labelBytes := buffer.Bytes()
	labelID := lib.MakeMDByByte(labelBytes)

	//labels index
	for _, v := range labels {
		//key type index:label:__name__#latency
		buffer := bytes.NewBufferString("index:label:")
		buffer.WriteString(v.Name)
		buffer.WriteString("#")
		buffer.WriteString(v.Value)
		key := buffer.String()
		//log.Println("Write label md:", key, labelID)

		//key type index:status:__name__#latency+labelID
		indexStatus := "index:status:" + v.Name + "#" + v.Value + "+" + labelID
		indexStatusKey, _ := tikv.Get([]byte(indexStatus))
		//log.Println("indexStatus:", indexStatusKey)

		//not in index
		if "" == indexStatusKey.Value {
			tikv.Puts([]byte(indexStatus), []byte("1"))

			//wtire tikv
			oldKey, _ := tikv.Get([]byte(key))
			if oldKey.Value == "" {
				tikv.Puts([]byte(key), []byte(labelID))
			} else {
				b := bytes.NewBufferString(oldKey.Value)
				b.WriteString(",")
				b.WriteString(labelID)
				v := b.Bytes()
				tikv.Puts([]byte(key), v)
			}
		}
	}

	tBuffer := bytes.NewBufferString("index:timeseries:")
	now := time.Now().UnixNano() / int64(time.Millisecond)
	now = (now / 300000) * 300000

	tBuffer.WriteString(labelID)
	tBuffer.WriteString(":")
	tBuffer.WriteString(strconv.FormatInt(now, 10))
	timeIndexBytes := tBuffer.Bytes()

	//samples index
	for _, v := range samples {
		oldKey, _ := tikv.Get(timeIndexBytes)
		if oldKey.Value == "" {
			tikv.Puts(timeIndexBytes, lib.Int64ToBytes(v.Timestamp))
		} else {
			b := bytes.NewBufferString(oldKey.Value)
			b.WriteString(",")
			b.Write(lib.Int64ToBytes(v.Timestamp))
			v := b.Bytes()
			tikv.Puts(timeIndexBytes, v)
		}
	}

	return labelID
}

func writeTimeseriesData(labelID string, samples []*prompb.Sample) {
	for _, v := range samples {
		//key type timeseries:doc:labelID#timestamp
		buffer := bytes.NewBufferString("timeseries:doc:")
		buffer.WriteString(labelID)
		buffer.WriteString(":")
		buffer.WriteString(strconv.FormatInt(v.Timestamp, 10))
		key := buffer.Bytes()
		//write to tikv
		tikv.Puts(key, []byte(strconv.FormatFloat(v.Value, 'E', -1, 64)))
		//log.Println("Write timeseries:", string(key), strconv.FormatFloat(v.Value, 'E', -1, 64))
	}
}

func SaveOriDoc(labelID string, originalMsg []byte) {
	buffer := bytes.NewBufferString("doc:")
	buffer.WriteString(labelID)
	key := buffer.Bytes()
	tikv.Puts(key, originalMsg)
	//log.Println("Write meta:", string(key), string(originalMsg))
}
