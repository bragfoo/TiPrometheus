package adapter

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/bragfoo/TiPrometheus/src/lib"
	"strconv"

	"../prompb"
	"../tikv"
	"bytes"
	"log"
	"sync"
	"time"
)

var metaLabelMap sync.Map

var indexLabelMap sync.Map

var oriMsgMap sync.Map

func RemoteWriter(data prompb.WriteRequest) {
	log.Println("tttttttttt")
	log.Println(data.Timeseries)
	for _, oneDoc := range data.Timeseries {
		labels := oneDoc.Labels
		samples := oneDoc.Samples
		//log.Println(labels, samples)

		//build index and return labelMD
		labelMD := buildIndex(labels, samples)

		// write
		writeTimeseriesData(labelMD, samples)
		//log.Println(metaLabelMap)
		labelsByte := lib.GetBytes(labels)
		SaveOriDoc(labelMD, labelsByte)
	}
	//time.Sleep(10*time.Minute)
	time.Sleep(10*time.Second)
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

	for _, v := range labels {
		//key type index:label:__name__#latency
		buffer := bytes.NewBufferString("index:label:")
		buffer.WriteString(v.Name)
		buffer.WriteString("#")
		buffer.WriteString(v.Value)

		key := buffer.String()
		//log.Println("Write label md:", key, labelMD)
		indexStatus := "index:status:" + v.Name + "#" + v.Value + "+" + labelMD
		indexStatusKey, gErr := tikv.Get([]byte(indexStatus))
		log.Println("indexStatus:", indexStatusKey, gErr)

		if "" == indexStatusKey.Value {
			tikv.Puts([]byte(indexStatus), []byte("1"))
			tikv.Puts([]byte(key), []byte(labelMD))
		} else {

			indexData, gErr := tikv.Get([]byte(key))
			log.Println("indexData:", indexData.Key, indexData.Value, gErr)
			//if "" == indexData.Value {
			//	tikv.Puts([]byte(key), []byte(labelMD))
			//} else {

			//tikv.Puts([]byte(indexStatus), []byte("1"))
			oldKey, err := tikv.Get([]byte(key))
			log.Println("oldKey", oldKey, err)
			b := bytes.NewBufferString(oldKey.Value)
			b.WriteString(",")
			b.WriteString(labelMD)
			v := b.Bytes()
			metaLabelMap.Store(key, string(v))
			tikv.Puts([]byte(key), v)
			//}

		}

		//tikv.Puts([]byte(key), []byte(labelMD))

		//if actual, loaded := metaLabelMap.LoadOrStore(key, labelMD); loaded {
		//	//insert value into old map value
		//	b := bytes.NewBufferString(actual.(string))
		//	b.WriteString(",")
		//	b.WriteString(labelMD)
		//	v := b.Bytes()
		//	metaLabelMap.Store(key, string(v))
		//	tikv.Puts([]byte(key), v)
		//} else {
		//	//重启 如果map里没有 需要直接写tikv
		//	//载入tikv中已有的信息
		//	if kv, err := tikv.Get([]byte(key)); "" != kv.Value && nil == err {
		//		b := bytes.NewBufferString(kv.Key)
		//		b.WriteString(",")
		//		b.WriteString(labelMD)
		//		v := b.Bytes()
		//		metaLabelMap.Store(key, string(v))
		//		tikv.Puts([]byte(key), v)
		//	}
		//	tikv.Puts([]byte(key), []byte(labelMD))
		//}

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
	for _, v := range samples {
		tikv.Puts(timeIndexBytes, int64ToBytes(v.Timestamp))
		//log.Println("Write time index:", string(timeIndexBytes), string(int64ToBytes(v.Timestamp)))

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

		//	//write to tikv
		tikv.Puts(key, []byte(strconv.FormatFloat(v.Value, 'E', -1, 64)))
		//log.Println("Write timeseries:", string(key), strconv.FormatFloat(v.Value, 'E', -1, 64))

	}
}

func SaveOriDoc(labelMD string, originalMsg []byte) {
	buffer := bytes.NewBufferString("doc:")
	buffer.WriteString(labelMD)
	key := buffer.Bytes()
	tikv.Puts(key, originalMsg)
	//log.Println("Write meta:", string(key), string(originalMsg))

}

func int64ToBytes(i int64) []byte {
	return []byte(strconv.FormatInt(i, 10))
}
