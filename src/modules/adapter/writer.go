package adapter

import (
	"crypto/md5"
	"encoding/hex"
	"strconv"
	
	"../prompb"
	"../tikv"
	"bytes"
	"encoding/binary"
	"log"
	"sync"
	"time"
)

var metaLabelMap sync.Map

var indexLabelMap sync.Map

func RemoteWriter(data prompb.WriteRequest) {
	for _, oneDoc := range data.Timeseries {
		labels := oneDoc.Labels
		samples := oneDoc.Samples
		log.Println(labels, samples)
		
		//build index and return labelMD
		labelMD := buildIndex(labels, samples)
		log.Println("LabelMD:", labelMD)

		// write
		writeTimeseriesData(labelMD, samples)
		log.Println(metaLabelMap)
	}
}

func MakeNewMD(initByte []byte) string {
	m := md5.New()
	m.Write(initByte)
	md := m.Sum(nil)
	mdString := hex.EncodeToString(md)
	return mdString
}


// build md5 data and store to kv if not exist
func buildIndex(labels []*prompb.Label, samples []*prompb.Sample) string {
	//make md
	var buffer = bytes.NewBufferString("")
	// key#value
	for _, v := range labels {
		buffer.WriteString(v.Name)
		buffer.WriteString("#")
		buffer.WriteString(v.Value)
	}
	labelBytes := buffer.Bytes()
	labelMD := MakeNewMD(labelBytes)

	//var sampleBuffer = bytes.NewBufferString("")
	//for _, v := range samples {
	//	sampleBuffer.WriteString(strconv.FormatFloat(v.Value, 'E', -1, 64))
	//	sampleBuffer.WriteString("#")
	//	sampleBuffer.WriteString(string(v.Timestamp))
	//}
	//sampleMD := MakeNewMD(sampleBuffer.Bytes())

	for _, v := range labels {
		//key type index:label:__name__#latency
		buffer := bytes.NewBufferString("index:label:")
		buffer.WriteString(v.Name)
		buffer.WriteString("#")
		buffer.WriteString(v.Value)
		//value
		//vBuffer := bytes.NewBufferString(labelMD)
		//vBuffer.WriteString(",")
		//vBuffer.WriteString(fmt.Sprintf("%s", sampleMD))

		key := buffer.String()
		//value := vBuffer.String()

		//finalKey := bytes.NewBufferString(key)
		//finalKey.WriteString(":")
		//finalKey.WriteString(value)
		//keyBytes := finalKey.Bytes()

		if actual, loaded := indexLabelMap.LoadOrStore(key, labelMD); loaded {
			//insert value into old map value
			b := bytes.NewBufferString(actual.(string))
			b.WriteString(",")
			b.WriteString(labelMD)
			v := b.Bytes()
			indexLabelMap.Store(key, string(v))
			tikv.Puts([]byte(key), v)
		} else {
			//重启 如果map里没有 需要直接写tikv
			// todo
			tikv.Puts([]byte(key), []byte(labelMD))
		}

		// debug info
		//indexLabelMap.Range(func(key, value interface{}) bool {
		//	log.Printf("[indexLabelMap]:[key]:%s [value]%s \n", key, value)
		//	return true
		//})

	}

	tBuffer := bytes.NewBufferString("index:timeseries:")
	now := time.Now().UnixNano() / int64(time.Millisecond)
	now = now / 300000
	now = now * 300000

	tBuffer.WriteString(labelMD)
	tBuffer.WriteString(":")
	tBuffer.WriteString(strconv.FormatInt(now, 10))
	timeIndexBytes := tBuffer.Bytes()
			//重启 如果map里没有 需要直接写tikv
			// todo
	for _, v := range samples {
		if _, ok := indexLabelMap.LoadOrStore(string(timeIndexBytes), v.Timestamp); !ok {
			tikv.Puts(timeIndexBytes, int64ToBytes(v.Timestamp))
		}
	}

	return labelMD
}

func writeTimeseriesData(labelMD string, samples []*prompb.Sample) {

	for _, v := range samples {
		buffer := bytes.NewBufferString("timeseries:doc:")
		buffer.WriteString(labelMD)
		buffer.WriteString(":")
		buffer.WriteString(strconv.FormatInt(v.Timestamp, 10))
		key := buffer.Bytes()
		// set cache map

		if _, ok := metaLabelMap.LoadOrStore(string(key), v.Value); !ok {
			//write to tikv
			tikv.Puts(key, []byte(strconv.FormatFloat(v.Value, 'E', -1, 64)))
		}
	}
}

func int64ToBytes(i int64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}
