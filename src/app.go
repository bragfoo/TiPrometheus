package main

import (
	"flag"
	"github.com/bragfoo/TiPrometheus/src/modules/process"
	"github.com/bragfoo/TiPrometheus/src/modules/simpleHTTP"
	"os"
)

// RunTime=dev go run app.go -conf "./conf/conf.toml"

var confPath = flag.String("conf", "./conf/conf.toml", "The conf path.")

func main() {
	flag.Parse()
	// init
	process.Init(os.Getenv("RunTime"), *confPath)
	// http server
	simpleHTTP.Server()
}
