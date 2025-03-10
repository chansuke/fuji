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
	"github.com/shiguredo/fuji/message"
)

type configWillTestCase struct {
	configStr     string          // testcase config file
	expectedError config.AnyError // expected error status
	message       string          // message when failed
}

var testcases = []configWillTestCase{
	// tests broker validation without will_message
	{
		configStr: `
                [[broker."sango/1"]]

                    host = "localhost"
                    port = 1883
`,
		expectedError: nil,
		message:       "WillMessage could not be omitted. Shall be optional."},
	// tests broker validation with will_message
	{
		configStr: `
                [[broker."sango/1"]]

                    host = "localhost"
                    port = 1883
		    will_message = "Hello world."
`,
		expectedError: nil,
		message:       "WillMessage could not be defined."},
	// tests broker validation with empty will_message
	{
		configStr: `
                [[broker."sango/1"]]

                    host = "localhost"
                    port = 1883
		    will_message = ""
`,
		expectedError: nil,
		message:       "Empty WillMessage could not be defined."},
	// tests multiple broker validation with only one will_message
	{
		configStr: `
                [[broker."sango/1"]]

                    host = "localhost"
                    port = 1883

                [[broker."sango/2"]]

                    host = "192.168.1.1 "
                    port = 1883
		    will_message = "Hello world."
`,
		expectedError: nil,
		message:       "WillMessage could not be defined for one of two."},
	// tests multiple broker validation with both will_message
	{
		configStr: `
                [[broker."sango/1"]]

                    host = "localhost"
                    port = 1883
		    will_message = "Change the world."

                [[broker."sango/2"]]

                    host = "192.168.1.1"
                    port = 1883
		    will_message = "Hello world."
`,
		expectedError: nil,
		message:       "WillMessage could not be defined for both of two."},
}

func generalConfigWillTest(test configWillTestCase, t *testing.T) {
	assert := assert.New(t)

	conf, err := config.LoadConfigByte([]byte(test.configStr))
	assert.Nil(err)

	brokers, err := broker.NewBrokers(conf, make(chan message.Message))
	assert.Nil(err)
	assert.NotEqual(0, len(brokers))
}

func TestConfigWillAll(t *testing.T) {
	for _, testcase := range testcases {
		generalConfigWillTest(testcase, t)
	}
}
