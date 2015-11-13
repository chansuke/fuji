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

func TestSearchBrokerSection(t *testing.T) {
	assert := assert.New(t)

	conf, err := LoadConfig("../tests/testing_conf.toml")
	assert.Nil(err)

	section := SearchSection(&conf.Sections, "broker", "1")
	assert.NotNil(section)

	section = SearchSection(&conf.Sections, "broker", "2")
	assert.NotNil(section)

	section = SearchSection(&conf.Sections, "broker", "3")
	assert.Nil(section)

}

func TestSearchDeviceType(t *testing.T) {
	assert := assert.New(t)

	conf, err := LoadConfig("../tests/testing_conf.toml")
	assert.Nil(err)

	section := SearchDeviceType(&conf.Sections, "serial")
	assert.NotNil(section)

	section = SearchDeviceType(&conf.Sections, "dummy")
	assert.NotNil(section)

	section = SearchDeviceType(&conf.Sections, "notfound")
	assert.Nil(section)

}
