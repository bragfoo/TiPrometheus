package adapter

import (
	"bytes"
	"encoding/gob"
	"github.com/bragfoo/TiPrometheus/pkg/conf"
	"github.com/bragfoo/TiPrometheus/pkg/lib"
	"github.com/bragfoo/TiPrometheus/pkg/tikv"
	"github.com/prometheus/prometheus/prompb"
	"go.uber.org/zap/buffer"
	"log"
	"math"
	"strconv"
)

var (
	pool = buffer.NewPool()
)

func RemoteReader(querys prompb.ReadRequest) *prompb.ReadResponse {
	//query
	query := querys.Queries[0]
	startTime := query.StartTimestampMs
	endTime := query.EndTimestampMs
	matchers := query.Matchers
	log.Println("Query:", startTime, endTime, matchers)

	//compute time endpoint
	timeEndpoinFromGet := getTimeEndpoint(startTime, endTime)

	//get data by matchers
	docTimeseries := getSameMatcher(matchers, timeEndpoinFromGet)

	//response
	var queryResult prompb.QueryResult
	queryResult.Timeseries = docTimeseries
	var queryResults []*prompb.QueryResult
	queryResults = append(queryResults, &queryResult)
	var resp prompb.ReadResponse
	resp.Results = queryResults

	return &resp
}

func getTimeEndpoint(startTime, endTime int64) []int64 {
	interval := float64(conf.RunTimeInfo.TimeInterval * 1000 * 60)
	startTimeCompute := (math.Floor(float64(startTime) / interval)) * interval
	endTimeCompute := (math.Floor(float64(endTime) / interval)) * interval
	//log.Println("Time compute:", int64(startTimeCompute), int64(endTimeCompute))

	//get timeEndpointList
	timeEndpointList := getTimeEndpointList(int64(startTimeCompute), int64(endTimeCompute), int64(interval))

	return timeEndpointList
}

func getTimeEndpointList(startTimeCompute, endTimeCompute, interval int64) []int64 {
	var timeEndpointList []int64
	//in one time interval
	if startTimeCompute == endTimeCompute {
		endTimeCompute = startTimeCompute + interval
	}
	//in time intervals
	timeEndpointList = append(timeEndpointList, int64(startTimeCompute))
	timeEndpoint := startTimeCompute
	for {
    use-interval-instead-of-hardcoded-value -- Incoming Change
		timeEndpoint = timeEndpoint + interval
		timeEndpointList = append(timeEndpointList, int64(timeEndpoint))
		if timeEndpoint == endTimeCompute {
			break
		}
	}
	//log.Println("Time endpoint list:", timeEndpointList)
	return timeEndpointList
}

func getSameMatcher(matchers []*prompb.LabelMatcher, timeEndpointList []int64) []*prompb.TimeSeries {
	buf := pool.Get()
	defer buf.Free()

	//get count map
	countMap := getCountMap(matchers)

	var docTimeseries []*prompb.TimeSeries

	//get intersection
	for md, count := range countMap {
		if count == len(matchers) {
			//log.Println("Find intersection key md:", md)

			//get labels info
			buf.AppendString("doc:")
			buf.AppendString(md)
			labelInfoKey := buf.Bytes()
			labelInfoKV, _ := tikv.Get([]byte(labelInfoKey))
			buf.Reset()

			//get labels
			labels := makeLabels([]byte(labelInfoKV.Value))

			//get timeseries list
			timeListString := getTimeList(md, timeEndpointList)
			timeList := lib.ReadStringByStepwidth(13, timeListString)

			//get values
			values := getValues(timeList, md)

			// one timeseries
			oneDocTimeseries := prompb.TimeSeries{
				Labels:  labels,
				Samples: values,
			}
			docTimeseries = append(docTimeseries, &oneDocTimeseries)
		}
	}

	log.Println("Response:", docTimeseries)
	return docTimeseries
}

func getCountMap(matchers []*prompb.LabelMatcher) map[string]int {
	buf := pool.Get()
	defer buf.Free()

	countMap := make(map[string]int)
	for _, queryLabel := range matchers {
		//newLabel
		buf.AppendString("index:label:")
		buf.AppendString(queryLabel.Name)
		buf.AppendString("#")
		buf.AppendString(queryLabel.Value)
		newLabel := buf.String()
		buf.Reset()

		//get label index list
		//key type index:label:newLabel
		newLabelValue, _ := tikv.Get([]byte(newLabel))
		mdList := lib.ReadStringByStepwidth(32, newLabelValue.Value)

		//mark count
		for _, oneMD := range mdList {
			oldCount := countMap[oneMD]
			newCount := oldCount + 1
			countMap[oneMD] = newCount
		}

	}

	//log.Println("Count Map:", countMap)
	return countMap
}

func makeLabels(labelInfoByte []byte) []*prompb.Label {
	var labels []*prompb.Label
	var buf bytes.Buffer
	// wtire to buffer
	buf.Write(labelInfoByte)
	dec := gob.NewDecoder(&buf)
	// read from buffer
	dec.Decode(&labels)
	//log.Println("Labels:", labels)
	return labels
}

func getTimeList(md string, timeEndpointList []int64) string {
	buf := pool.Get()
	defer buf.Free()

	var timeList string
	//key type index:timeseries:5d4decf2a1d0dd0151cd893cfc752af4:1543639500000
	for _, oneTimeEndpoint := range timeEndpointList {
		buf.AppendString("index:timeseries:")
		buf.AppendString(md)
		buf.AppendString(":")
		buf.AppendString(strconv.FormatInt(oneTimeEndpoint, 10))
		timeIndexBytes := buf.Bytes()
		buf.Reset()
		newLabelValue, _ := tikv.Get(timeIndexBytes)
		//log.Println("One time endpoint list:", newLabelValue.Value)

		if newLabelValue.Value != "" {
			//split with ,
			//timeList = append(timeList, strings.Split(newLabelValue.Value, ",")...)
			timeList = timeList + newLabelValue.Value
		}

	}

	//log.Println("Time list:", timeList)
	return timeList
}

func getValues(timeList []string, md string) []prompb.Sample {
	var values []prompb.Sample

	bvChan := make(chan prompb.Sample, 1000)

	for _, oneTimePoint := range timeList {
		//get time point and value
		go getTimePointValue(oneTimePoint, md, bvChan)
	}

	// init count
	var bvNum int
	// read from channel
	for {
		baseValue := <-bvChan
		bvNum = bvNum + 1
		values = append(values, baseValue)
		if bvNum == len(timeList) {
			close(bvChan)
			break
		}
	}

	return values
}

func getTimePointValue(oneTimePoint, md string, bvChan chan prompb.Sample) {
	buf := pool.Get()
	//key type timeseries:doc:5d4decf2a1d0dd0151cd893cfc752af4:1543639730686
	buf.AppendString("timeseries:doc:")
	buf.AppendString(md)
	buf.AppendString(":")
	buf.AppendString(oneTimePoint)
	key := buf.Bytes()
	oneTimePointValue, _ := tikv.Get([]byte(key))
	buf.Free()
	//log.Println("One doc:", oneTimeseriesValue)

	oneTimePointValueFloat, _ := strconv.ParseFloat(oneTimePointValue.Value, 64)
	oneTimePointInt, _ := strconv.ParseInt(oneTimePoint, 10, 64)

	baseValue := prompb.Sample{
		Value:     oneTimePointValueFloat,
		Timestamp: oneTimePointInt,
	}

	bvChan <- baseValue
}
