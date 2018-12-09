package adapter

import (
	"../prompb"
	"../tikv"
	"bytes"
	"encoding/gob"
	"log"
	"math"
	"strconv"
	"strings"
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
	startTimeCompute := (math.Floor(float64(startTime) / 300000)) * 300000
	endTimeCompute := (math.Floor(float64(endTime) / 300000)) * 300000
	//log.Println("Time compute:", int64(startTimeCompute), int64(endTimeCompute))

	var tiemEndpointList []int64
	//in one time interval
	if startTimeCompute == endTimeCompute {
		endTimeCompute = startTimeCompute + 300000
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

	//log.Println("Time endpoint:", tiemEndpointList)
	return tiemEndpointList
}

func getSameMatcher(matchers []*prompb.LabelMatcher, tiemEndpointList []int64) []*prompb.TimeSeries {
	countMap := make(map[string]int)
	var docTimeseries []*prompb.TimeSeries

	for _, queryLabel := range matchers {
		//newLabel
		buffer := bytes.NewBufferString("index:label:")
		buffer.WriteString(queryLabel.Name)
		buffer.WriteString("#")
		buffer.WriteString(queryLabel.Value)
		newLabel := buffer.String()
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

	//get same md
	for md, count := range countMap {
		//in the same doc
		if count == len(matchers) {
			//log.Println("Find intersection key md:", md)

			//get labels info
			buffer := bytes.NewBufferString("doc:")
			buffer.WriteString(md)
			labelInfoKey := buffer.Bytes()
			labelInfoKV, _ := tikv.Get([]byte(labelInfoKey))
			//log.Println("One label info", labelInfoKV)

			//get labels
			labels := makeLabels([]byte(labelInfoKV.Value))

			//get values
			var values []*prompb.Sample
			//get timeseries list
			timeList := getTimeList(md, tiemEndpointList)
			for _, oneTimeseries := range timeList {
				//key type timeseries:doc:5d4decf2a1d0dd0151cd893cfc752af4:1543639730686
				buffer := bytes.NewBufferString("timeseries:doc:")
				buffer.WriteString(md)
				buffer.WriteString(":")
				buffer.WriteString(oneTimeseries)
				key := buffer.Bytes()
				oneTimeseriesValue, _ := tikv.Get([]byte(key))
				//log.Println("One doc value:", oneTimeseriesValue)

				//make value
				oneTimeseriesValueFloat, _ := strconv.ParseFloat(oneTimeseriesValue.Value, 64)
				oneTimeseriesInt, _ := strconv.ParseInt(oneTimeseries, 10, 64)
				baseValue := prompb.Sample{
					Value:     oneTimeseriesValueFloat,
					Timestamp: oneTimeseriesInt,
				}
				values = append(values, &baseValue)
			}

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
			//log.Println("One time list:", newLabelValue)
			timeListTemp := newLabelValue.Value
			timeList = append(timeList, strings.Split(timeListTemp, ",")...)
		}
	}
	
	//log.Println("Time list:", timeList)
	return timeList
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
