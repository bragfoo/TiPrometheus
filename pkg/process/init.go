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

package process

import (
	"log"

	"github.com/BurntSushi/toml"
	"github.com/bragfoo/TiPrometheus/pkg/conf"
	"github.com/bragfoo/TiPrometheus/pkg/lib"
	"github.com/bragfoo/TiPrometheus/pkg/tikv"
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
