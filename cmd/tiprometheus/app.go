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

package main

import (
	"flag"
	"os"

	"github.com/bragfoo/TiPrometheus/pkg/conf"
	"github.com/bragfoo/TiPrometheus/pkg/process"
	"github.com/bragfoo/TiPrometheus/pkg/simpleHTTP"
)

// RunTime=dev go run app.go -conf "./conf.toml"

var confPath = flag.String("conf", "./conf.toml", "The configuration file name.")

func main() {
	flag.Parse()
	// init
	process.Init(os.Getenv("RunTime"), *confPath)
	if conf.RunTimeInfo.AdapterEnableTLS {
		// start https server
		simpleHTTP.ServerTLS(conf.RunTimeInfo.AdapterListen, conf.RunTimeInfo.AdapterCACertificate, conf.RunTimeInfo.AdapterServerCertificate, conf.RunTimeInfo.AdapterServerKey)
	} else {
		// start http server
		simpleHTTP.Server(conf.RunTimeInfo.AdapterListen)
	}
}
