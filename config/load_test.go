// Copyright 2015 Shiguredo Inc. <fuji@shiguredo.jp>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDupGateway(t *testing.T) {
	assert := assert.New(t)

	configStr := `
[[gateway]]
    name = "ham"
    max_retry_count = 30
`
	_, err := LoadConfigByte([]byte(configStr))
	assert.NotNil(err)

}

func TestUniqBroker(t *testing.T) {
	assert := assert.New(t)

	configStr := `
[broker."sango/2"]
    host = "192.168.1.22"
    port = 1883
`
	_, err := LoadConfigByte([]byte(configStr))
	assert.NotNil(err)

}

func TestDupDevice(t *testing.T) {
	assert := assert.New(t)

	configStr := `
[[device."dora/dummy"]]
    broker = "sango"
    qos = 1
    dummy = true
    interval = 10
    payload = "Hello world."
`
	_, err := LoadConfigByte([]byte(configStr))
	assert.NotNil(err)

}
