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
	if conf.RunTimeInfo.AdapterEnableTLS {
		// start https server
		simpleHTTP.ServerTLS(conf.RunTimeInfo.AdapterListen, conf.RunTimeInfo.AdapterCACertificate, conf.RunTimeInfo.AdapterServerCertificate, conf.RunTimeInfo.AdapterServerKey)
	} else {
		// start http server
		simpleHTTP.Server(conf.RunTimeInfo.AdapterListen)
	}
}
