package adapter

import (
	"github.com/bragfoo/TiPrometheus/pkg/conf"
	"github.com/bragfoo/TiPrometheus/pkg/lib"
	"log"
	"strconv"

	"bytes"
	"github.com/bragfoo/TiPrometheus/pkg/tikv"
	"github.com/prometheus/prometheus/prompb"
	"go.uber.org/zap/buffer"
	"time"
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

		labelsByte := lib.GetBytes(labels)
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
			tikv.Puts([]byte(indexStatus), []byte("1"))

			//wtire tikv
			oldKey, _ := tikv.Get([]byte(key))
			if oldKey.Value == "" {
				tikv.Puts([]byte(key), []byte(labelID))
			} else {
				b := bytes.NewBufferString(oldKey.Value)
				b.WriteString(labelID)
				v := b.Bytes()
				tikv.Puts([]byte(key), v)
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
			tikv.Puts(timeIndexBytes, lib.Int64ToBytes(v.Timestamp))
		} else {
			bs := buffers.Get()
			bs.AppendString(oldKey.Value)
			bs.AppendString(strconv.FormatInt(v.Timestamp, 10))
			v := bs.Bytes()
			tikv.Puts(timeIndexBytes, v)
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
		tikv.Puts(key, []byte(strconv.FormatFloat(v.Value, 'E', -1, 64)))
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
	tikv.Puts(key, originalMsg)
	//log.Println("Write meta:", string(key), string(originalMsg))
}
