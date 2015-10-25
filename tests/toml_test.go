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

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/shiguredo/fuji/broker"
	"github.com/shiguredo/fuji/device"
	"github.com/shiguredo/fuji/gateway"
	"github.com/shiguredo/fuji/message"
	"github.com/shiguredo/fuji/toml"
)

func TestIniLoadini(t *testing.T) {
	assert := assert.New(t)

	_, err := toml.LoadConfig("testing_conf.toml")
	assert.Nil(err)
}

func TestIniNewGateway(t *testing.T) {
	assert := assert.New(t)

	conf, err := toml.LoadConfig("testing_conf.toml")
	assert.Nil(err)
	gw, err := gateway.NewGateway(conf)
	assert.Nil(err)
	assert.Equal("ham", gw.Name)
}

func TestIniNewBrokers(t *testing.T) {
	assert := assert.New(t)

	conf, err := toml.LoadConfig("testing_conf.toml")
	assert.Nil(err)
	brokerList, err := broker.NewBrokers(conf, make(chan message.Message))
	assert.Nil(err)
	assert.Equal(3, len(brokerList))
}

func TestIniNewSerialDevices(t *testing.T) {
	assert := assert.New(t)

	conf, err := toml.LoadConfig("testing_conf.toml")
	brokerList, err := broker.NewBrokers(conf, make(chan message.Message))
	assert.Nil(err)
	deviceList, _, err := device.NewDevices(conf, brokerList)
	assert.Nil(err)
	assert.Equal(3, len(deviceList))
}

func TestIniNewDummyDevice(t *testing.T) {
	assert := assert.New(t)

	conf, err := toml.LoadConfig("testing_conf.toml")
	brokerList, err := broker.NewBrokers(conf, make(chan message.Message))
	assert.Nil(err)

	dummy, err := device.NewDummyDevice(conf.Sections[7], brokerList, device.NewDeviceChannel())
	assert.Nil(err)
	assert.Equal("dummy", dummy.DeviceType())
	assert.Equal(2, int(dummy.QoS))
}
