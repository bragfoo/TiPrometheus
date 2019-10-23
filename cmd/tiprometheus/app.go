package main

import (
	"flag"
	"github.com/bragfoo/TiPrometheus/pkg/process"
	"github.com/bragfoo/TiPrometheus/pkg/conf"
	"github.com/bragfoo/TiPrometheus/pkg/simpleHTTP"
	"os"
)

// RunTime=dev go run app.go -conf "./conf.toml"

var confPath = flag.String("conf", "./conf.toml", "The configuration file name.")

func main() {
	flag.Parse()
	// init
	process.Init(os.Getenv("RunTime"), *confPath)
		// start http server
		simpleHTTP.Server(conf.RunTimeInfo.AdapterListen)
}
