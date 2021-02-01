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

package conf

// AgentConf is the agent configuration type.
type AgentConf struct {
	PDHost                   string
	TimeInterval             int
	AdapterListen            string
	AdapterEnableTLS         bool
	AdapterCACertificate     string
	AdapterServerCertificate string
	AdapterServerKey         string
	TiKVEnableTLS            bool
	TiKVCACertificate        string
	TiKVClientCertificate    string
	TiKVClientKey            string
}

// RunTimeInfo contains the active configuration.
var RunTimeInfo AgentConf

// RunTimeMap contains the full configuration file.
var RunTimeMap map[string]AgentConf

// DefaultRunTimeName specifies the name of the default TOML section
// to be loaded when nothing is configured.
const DefaultRunTimeName = "default"
