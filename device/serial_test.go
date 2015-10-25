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
	"github.com/shiguredo/fuji/toml"
)

func TestNewSerialDevice(t *testing.T) {
	assert := assert.New(t)

	iniStr := `
[[device."dora/serial"]]
    broker = "sango"
    qos = 1
    serial = "/dev/tty.ble"
    baud = 9600
    size = 4
    type = "BLE"
`
	conf, err := toml.LoadConfigByte([]byte(iniStr))
	b1 := &broker.Broker{Name: "sango"}
	brokers := []*broker.Broker{b1}
	b, err := NewSerialDevice(conf.Sections[0], brokers, NewDeviceChannel())
	assert.Nil(err)
	assert.NotNil(b.Broker)
	assert.Equal("dora", b.Name)
	assert.Equal(byte(1), b.QoS)
	assert.Equal(4, b.Size)
	assert.Equal("BLE", b.Type)
}

func TestNewSerialDeviceNotSetSize(t *testing.T) {
	assert := assert.New(t)

	iniStr := `
[[device."dora/serial"]]
    broker = "sango"
    qos = 1
    serial = "/dev/tty.ble"
    baud = 9600
    type = "BLE"
`
	conf, err := toml.LoadConfigByte([]byte(iniStr))
	b1 := &broker.Broker{Name: "sango"}
	brokers := []*broker.Broker{b1}
	b, err := NewSerialDevice(conf.Sections[0], brokers, NewDeviceChannel())
	assert.Nil(err)
	assert.NotNil(b.Broker)
	assert.Equal("dora", b.Name)
	assert.Equal(0, b.Size)
	assert.Equal("BLE", b.Type)
}

func TestNewSerialDeviceInvalidInterval(t *testing.T) {
	assert := assert.New(t)

	iniStr := `
[[device."dora/serial"]]
    broker = "sango"
    interval = -1
    qos = 1
`
	conf, err := toml.LoadConfigByte([]byte(iniStr))
	b1 := &broker.Broker{Name: "sango"}
	brokers := []*broker.Broker{b1}
	_, err = NewSerialDevice(conf.Sections[0], brokers, NewDeviceChannel())
	assert.NotNil(err)
}

func TestNewSerialDeviceInvalidQoS(t *testing.T) {
	assert := assert.New(t)

	iniStr := `
[[device."dora/serial"]]
    broker = "sango"
    qos = -1
`
	conf, err := toml.LoadConfigByte([]byte(iniStr))
	b1 := &broker.Broker{Name: "sango"}
	brokers := []*broker.Broker{b1}
	_, err = NewSerialDevice(conf.Sections[0], brokers, NewDeviceChannel())
	assert.NotNil(err)
}

func TestNewSerialDeviceInvalidBroker(t *testing.T) {
	assert := assert.New(t)

	iniStr := `
[[device."dora/serial"]]
    broker = "doesNotExist"
    qos = 1
`
	conf, err := toml.LoadConfigByte([]byte(iniStr))
	b1 := &broker.Broker{Name: "sango"}
	brokers := []*broker.Broker{b1}
	_, err = NewSerialDevice(conf.Sections[0], brokers, NewDeviceChannel())
	assert.NotNil(err)
}

func TestNewSerialDeviceInvalidBaud(t *testing.T) {
	assert := assert.New(t)

	iniStr := `
[[device."dora/serial"]]
    broker = "sango"
    qos = -1
    baud = -9600
`
	conf, err := toml.LoadConfigByte([]byte(iniStr))
	b1 := &broker.Broker{Name: "sango"}
	brokers := []*broker.Broker{b1}
	_, err = NewSerialDevice(conf.Sections[0], brokers, NewDeviceChannel())
	assert.NotNil(err)
}

// TODO: TestIniBadDeviceWithUnknownInterface
