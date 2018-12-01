package process

import (
	"log"

	"../../lib"
	"../conf"
	"../tikv"

	"github.com/BurntSushi/toml"
)

// Init is init data
func Init(runTime, confPath string) {
	// init runtime
	if _, err := toml.DecodeFile(confPath, &conf.RunTimeMap); err != nil {
		log.Println(err)
		return
	}
	conf.RunTimeInfo = conf.RunTimeMap[runTime]
	log.Println(conf.RunTimeMap)
	// init log
	lib.InitLog()
	// init tikv
	tikv.InitStore()
}
