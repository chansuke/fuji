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
	"github.com/shiguredo/fuji/config"
	"github.com/shiguredo/fuji/device"
	"github.com/shiguredo/fuji/message"
)

// configRetainTestCase はRetain機能のテストの条件を示すデータ型です。
// configString は設定ファイルの内容
// expectedError はテストを実行したときに期待されるエラーの状態
// message はテストが失敗した内容の説明
type configRetainTestCase struct {
	configStr     string
	expectedError config.AnyError
	message       string
}

var serialDeviceTestcases = []configRetainTestCase{
	// check device validation without retain flag
	{
		configStr: `
		[[broker."sango/1"]]
		host = "localhost"
		port = 1883

		[device."hi"]
		type = "serial"
		broker = "sango"
		serial = "/dev/tty"
		baud = 9600
		qos = 0
`,
		expectedError: nil,
		message:       "Retain flag could not be omitted. Shall be optional."},
	// check device validation with retain flag
	{
		configStr: `
		[[broker."sango/1"]]
		host = "localhost"
		port = 1883

		[device."hi"]
		type = "serial"
		broker = "sango"
		serial = "/dev/tty"
		baud = 9600
		qos = 0
		retain = true
`,
		expectedError: nil,
		message:       "Retain flag could not be set."},
	// check device validation with retain flag is false
	{
		configStr: `
		[[broker."sango/1"]]
		host = "localhost"
		port = 1883

		[device."hi"]
		type = "serial"
		broker = "sango"
		serial = "/dev/tty"
		baud = 9600
		qos = 0
		retain = false 
`,
		expectedError: nil,
		message:       "Retain flag could not be un-set."},
}

var dummyDeviceTestcases = []configRetainTestCase{
	// check device validation without retain flag
	{
		configStr: `
		[[broker."sango/1"]]
		host = "localhost"
		port = 1883

		[device."hi"]
		type = "dummy"
		broker = "sango"
		qos = 0
		interval = 10
		payload = "Hello world."
`,
		expectedError: nil,
		message:       "Retain flag could not be omitted. Shall be optional."},
	// check device validation with retain flag
	{
		configStr: `
		[[broker."sango/1"]]
		host = "localhost"
		port = 1883

		[device."hi"]
		type = "dummy"
		broker = "sango"
		qos = 0
		retain = true
		interval = 10
		payload = "Hello world."
`,
		expectedError: nil,
		message:       "Retain flag could not be set."},
	// check device validation with retain flag is false
	{
		configStr: `
		[[broker."sango/1"]]
		host = "localhost"
		port = 1883

                [device."hi"]
		type = "dummy"
		broker = "sango"
		qos = 0
		retain = false 
		interval = 10
		payload = "Hello world."
`,
		expectedError: nil,
		message:       "Retain flag could not be un-set."},
}

// generalConfigRetainSerialDeviceTest checks retain function with serial device
func generalConfigRetainSerialDeviceTest(test configRetainTestCase, t *testing.T) {
	assert := assert.New(t)

	conf, err := config.LoadConfigByte([]byte(test.configStr))
	assert.Nil(err)

	brokers, err := broker.NewBrokers(conf, make(chan message.Message))
	assert.Nil(err)

	devices, _, err := device.NewDevices(conf, brokers)
	assert.Nil(err)
	assert.Equal(1, len(devices))
}

// generalConfigRetainDummyDeviceTest checks retain function with dummy device
func generalConfigRetainDummyDeviceTest(test configRetainTestCase, t *testing.T) {
	assert := assert.New(t)

	conf, err := config.LoadConfigByte([]byte(test.configStr))
	assert.Nil(err)

	brokers, err := broker.NewBrokers(conf, make(chan message.Message))
	assert.Nil(err)

	dummy, err := device.NewDummyDevice(conf.Sections[1], brokers, device.NewDeviceChannel())
	if test.expectedError == nil {
		assert.Nil(err)
		assert.NotNil(dummy)
	} else {
		assert.NotNil(err)
	}
}

// TestConfigRetainDeviceAll tests a serial device using test code
func TestConfigRetainDeviceAll(t *testing.T) {
	i := 0
	for _, testcase := range serialDeviceTestcases {
		generalConfigRetainSerialDeviceTest(testcase, t)
		i++
	}
}

// TestConfigRetainDeviceAll tests a dummy device using test code
func TestConfigRetainDummyDeviceAll(t *testing.T) {
	for _, testcase := range dummyDeviceTestcases {
		generalConfigRetainDummyDeviceTest(testcase, t)
	}
}
