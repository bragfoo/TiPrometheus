package process

import (
	"github.com/BurntSushi/toml"
	"github.com/bragfoo/TiPrometheus/pkg/conf"
	"github.com/bragfoo/TiPrometheus/pkg/lib"
	"github.com/bragfoo/TiPrometheus/pkg/tikv"
	"log"
)

// Init loads the configuration file and initializes
// logging and other subsystems.
func Init(runTime, confPath string) {
	// load the config file
	if _, err := toml.DecodeFile(confPath, &conf.RunTimeMap); err != nil {
		log.Println(err)
		return
	}
	// fall back to the default config section when unspecified
	if runTime == "" {
		runTime = conf.DefaultRunTimeName
	}
	conf.RunTimeInfo = conf.RunTimeMap[runTime]
	log.Println(conf.RunTimeInfo)
	// init log
	lib.InitLog()
	// init tikv client lib
	if conf.RunTimeInfo.TiKVEnableTLS {
		tikv.Init([]string{conf.RunTimeInfo.PDHost}, conf.RunTimeInfo.TiKVCACertificate, conf.RunTimeInfo.TiKVClientCertificate, conf.RunTimeInfo.TiKVClientKey)
	} else {
		tikv.Init([]string{conf.RunTimeInfo.PDHost}, "", "", "")
	}
}
