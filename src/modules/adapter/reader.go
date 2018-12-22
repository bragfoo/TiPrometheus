package adapter

import (
	"bytes"
	"encoding/gob"
	"github.com/bragfoo/TiPrometheus/src/modules/conf"
	"github.com/bragfoo/TiPrometheus/src/modules/prompb"
	"github.com/bragfoo/TiPrometheus/src/modules/tikv"
	"go.uber.org/zap/buffer"
	"log"
	"math"
	"strconv"
	"strings"
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
	tiemEndpoinFromGet := getTimeEndpoint(startTime, endTime)

	//get data by matchers
	docTimeseries := getSameMatcher(matchers, tiemEndpoinFromGet)

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

	//get tiemEndpointList
	tiemEndpointList := getTiemEndpointList(int64(startTimeCompute), int64(endTimeCompute), int64(interval))

	return tiemEndpointList
}

func getTiemEndpointList(startTimeCompute, endTimeCompute, interval int64) []int64 {
	var tiemEndpointList []int64
	//in one time interval
	if startTimeCompute == endTimeCompute {
		endTimeCompute = startTimeCompute + interval
	}
	//in time intervals
	tiemEndpointList = append(tiemEndpointList, int64(startTimeCompute))
	tiemEndpoint := startTimeCompute
	for {
		tiemEndpoint = tiemEndpoint + 300000
		tiemEndpointList = append(tiemEndpointList, int64(tiemEndpoint))
		if tiemEndpoint == endTimeCompute {
			break
		}
	}
	log.Println("Time endpoint list:", tiemEndpointList)
	return tiemEndpointList
}

func getSameMatcher(matchers []*prompb.LabelMatcher, tiemEndpointList []int64) []*prompb.TimeSeries {
	buffer := pool.Get()
	defer buffer.Free()

	//get count map
	countMap := getCountMap(matchers, buffer)

	var docTimeseries []*prompb.TimeSeries

	//get intersection
	for md, count := range countMap {
		if count == len(matchers) {
			//log.Println("Find intersection key md:", md)

			//get labels info
			buffer.AppendString("doc:")
			buffer.AppendString(md)
			labelInfoKey := buffer.Bytes()
			labelInfoKV, _ := tikv.Get([]byte(labelInfoKey))
			buffer.Reset()

			//get labels
			labels := makeLabels([]byte(labelInfoKV.Value))

			//get timeseries list
			timeList := getTimeList(md, tiemEndpointList)

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

func getCountMap(matchers []*prompb.LabelMatcher, buffer *buffer.Buffer) map[string]int {
	countMap := make(map[string]int)
	for _, queryLabel := range matchers {
		//newLabel
		buffer.AppendString("index:label:")
		buffer.AppendString(queryLabel.Name)
		buffer.AppendString("#")
		buffer.AppendString(queryLabel.Value)
		newLabel := buffer.String()
		buffer.Reset()

		//get label index list
		//key type index:label:newLabel
		newLabelValue, _ := tikv.Get([]byte(newLabel))
		mdList := strings.Split(newLabelValue.Value, ",")

		//mark count
		for _, oneMD := range mdList {
			oldCount := countMap[oneMD]
			newCount := oldCount + 1
			countMap[oneMD] = newCount
		}
	}

	log.Println("Count Map:", countMap)
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

func getTimeList(md string, tiemEndpointList []int64) []string {
	var timeList []string
	//key type index:timeseries:5d4decf2a1d0dd0151cd893cfc752af4:1543639500000
	for _, oneTimeEndpoint := range tiemEndpointList {
		buffer := bytes.NewBufferString("index:timeseries:")
		buffer.WriteString(md)
		buffer.WriteString(":")
		buffer.WriteString(strconv.FormatInt(oneTimeEndpoint, 10))
		timeIndexBytes := buffer.Bytes()
		newLabelValue, _ := tikv.Get(timeIndexBytes)
		if newLabelValue.Value != "" {
			timeListTemp := newLabelValue.Value
			timeList = append(timeList, strings.Split(timeListTemp, ",")...)
		}
	}

	log.Println("Time list:", timeList)
	return timeList
}

func getValues(timeList []string, md string) []*prompb.Sample {
	var values []*prompb.Sample
	
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
		values = append(values, &baseValue)
		if bvNum == len(timeList) {
			close(bvChan)
			break
		}
	}
	
	return values
}

func getTimePointValue(oneTimePoint, md string, bvChan chan prompb.Sample) {
	buffer := pool.Get()
	//key type timeseries:doc:5d4decf2a1d0dd0151cd893cfc752af4:1543639730686
	buffer.AppendString("timeseries:doc:")
	buffer.AppendString(md)
	buffer.AppendString(":")
	buffer.AppendString(oneTimePoint)
	key := buffer.Bytes()
	oneTimePointValue, _ := tikv.Get([]byte(key))
	buffer.Free()
	//log.Println("One doc:", oneTimeseriesValue)

	oneTimePointValueFloat, _ := strconv.ParseFloat(oneTimePointValue.Value, 64)
	oneTimePointInt, _ := strconv.ParseInt(oneTimePoint, 10, 64)

	baseValue := prompb.Sample{
		Value:     oneTimePointValueFloat,
		Timestamp: oneTimePointInt,
	}

	bvChan <- baseValue
}
