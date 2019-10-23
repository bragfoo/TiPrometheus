package process

import (
	"github.com/BurntSushi/toml"
	"github.com/bragfoo/TiPrometheus/pkg/lib"
	"github.com/bragfoo/TiPrometheus/pkg/conf"
	"github.com/bragfoo/TiPrometheus/pkg/tikv"
	"log"
)

// Init is init data
func Init(runTime, confPath string) {
	// init runtime
	if _, err := toml.DecodeFile(confPath, &conf.RunTimeMap); err != nil {
		log.Println(err)
		return
	}
	conf.RunTimeInfo = conf.RunTimeMap[runTime]
	log.Println(conf.RunTimeInfo)
	// init log
	lib.InitLog()
	// init tikv
	tikv.Init()
}
