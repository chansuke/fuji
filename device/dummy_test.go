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

package device

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/shiguredo/fuji/broker"
	"github.com/shiguredo/fuji/config"
)

func TestNewDummyDevice(t *testing.T) {
	assert := assert.New(t)

	iniStr := `
[[device."dora/dummy"]]
    broker = "sango"
    qos = 1
    dummy = true
    interval = 10
    payload = "Hello world."
`
	conf, err := config.LoadConfigByte([]byte(iniStr))
	assert.Nil(err)
	b1 := &broker.Broker{Name: "sango"}
	brokers := []*broker.Broker{b1}
	b, err := NewDummyDevice(conf.Sections[0], brokers, NewDeviceChannel())
	assert.Nil(err)
	assert.NotNil(b.Broker)
	assert.Equal("dora", b.Name)
	assert.Equal(byte(1), b.QoS)
	assert.Equal(10, b.Interval)
	assert.Equal([]byte("Hello world."), b.Payload)
}

func TestNewDummyDeviceInvalidInterval(t *testing.T) {
	assert := assert.New(t)

	iniStr := `
[[device."dora/dummy"]]
    broker = "sango"
    interval = -1
    qos = 1
`
	conf, err := config.LoadConfigByte([]byte(iniStr))
	assert.Nil(err)
	b1 := &broker.Broker{Name: "sango"}
	brokers := []*broker.Broker{b1}
	_, err = NewDummyDevice(conf.Sections[0], brokers, NewDeviceChannel())
	assert.NotNil(err)
}

func TestNewDummyDeviceInvalidQoS(t *testing.T) {
	assert := assert.New(t)

	iniStr := `
[[device."dora/dummy"]]
    broker = "sango"
    interval = -1
    qos = -1
`
	conf, err := config.LoadConfigByte([]byte(iniStr))
	assert.Nil(err)
	b1 := &broker.Broker{Name: "sango"}
	brokers := []*broker.Broker{b1}
	_, err = NewDummyDevice(conf.Sections[0], brokers, NewDeviceChannel())
	assert.NotNil(err)
}

func TestNewDummyDeviceInvalidBroker(t *testing.T) {
	assert := assert.New(t)

	iniStr := `
[[device."dora/dummy"]]
    broker = "doesNotExist"
    interval = 10
    qos = 1
`
	conf, err := config.LoadConfigByte([]byte(iniStr))
	assert.Nil(err)
	b1 := &broker.Broker{Name: "sango"}
	brokers := []*broker.Broker{b1}
	_, err = NewDummyDevice(conf.Sections[0], brokers, NewDeviceChannel())
	assert.NotNil(err)
}
