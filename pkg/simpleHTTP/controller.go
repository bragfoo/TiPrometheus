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

package simpleHTTP

import (
	"io/ioutil"
	"net/http"

	"github.com/bragfoo/TiPrometheus/pkg/adapter"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/prometheus/prometheus/prompb"
)

func RemoteWrite(w http.ResponseWriter, r *http.Request) {
	compressed, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// resolve snappy
	reqBuf, err := snappy.Decode(nil, compressed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// resolve json
	var wreq prompb.WriteRequest

	if err := proto.Unmarshal(reqBuf, &wreq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// resolve data
	adapter.RemoteWriter(wreq)
	if _, err := w.Write([]byte("ok")); err != nil {
		return
	}
}

func RemoteRead(w http.ResponseWriter, r *http.Request) {
	compressed, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// snappy
	reqBuf, err := snappy.Decode(nil, compressed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// resolve json
	var rreq prompb.ReadRequest
	if err := proto.Unmarshal(reqBuf, &rreq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	naiveData := adapter.RemoteReader(rreq)
	data, _ := proto.Marshal(naiveData)
	// sender
	w.Header().Set("Content-Type", "application/x-protobuf")
	w.Header().Set("Content-Encoding", "snappy")
	compressed = snappy.Encode(nil, data)
	if _, err := w.Write(compressed); err != nil {
		return
	}
}
