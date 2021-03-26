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

package lib

import (
	"log"
	"os"
	"strings"
	"time"
)

// InitLog is init log info
func InitLog() {
	log.SetOutput(os.Stdout)
}

// CustomLogger is custom log for self
func CustomLogger(args ...string) {
	logInfo := strings.Join(args, " ")
	timeMark := time.Now()
	log.Println(timeMark.Format("[2006/01/02 15:04:05]"), logInfo)
}
