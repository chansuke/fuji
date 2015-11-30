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

package broker

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/shiguredo/fuji/config"
	"github.com/shiguredo/fuji/message"
)

/*
func TestNewBrokersSingle(t *testing.T) {
	assert := assert.New(t)

	configStr := `
[[broker."sango/2"]]
    host = "192.168.1.22"
    port = 1883
`
	conf, err := config.LoadConfigByte([]byte(configStr))
	b, err := NewBrokers(conf, make(chan message.Message))
	assert.Nil(err)
	assert.Equal(1, len(b))
	assert.Equal("sango", b[0].Name)
	assert.Equal(2, b[0].Priority)
	assert.Equal("", b[0].TopicPrefix)
	assert.Equal([]byte{}, b[0].WillMessage)
}
*/
func TestNewBrokersSettings(t *testing.T) {
	assert := assert.New(t)

	configStr := `
[[broker."sango/2"]]
    host = "192.168.1.22"
    port = 1883
    username = "usr"
    password = "pass"
    topic_prefix = "pre"
    will_message = "will"
`
	conf, err := config.LoadConfigByte([]byte(configStr))
	b, err := NewBrokers(conf, make(chan message.Message))
	assert.Nil(err)
	assert.Equal(1, len(b))
	assert.Equal("usr", b[0].Username)
	assert.Equal("pass", b[0].Password)
	assert.Equal("pre", b[0].TopicPrefix)
	assert.Equal([]byte("will"), b[0].WillMessage)
}

func TestNewBrokersMulti(t *testing.T) {
	assert := assert.New(t)

	configStr := `
[[broker."sango/1"]]
    host = "192.168.1.22"
    port = 1883
[[broker."sango/2"]]
    host = "192.168.1.22"
    port = 1883
`
	conf, err := config.LoadConfigByte([]byte(configStr))
	b, err := NewBrokers(conf, make(chan message.Message))
	assert.Nil(err)
	assert.Equal(2, len(b))
	assert.Equal("sango", b[0].Name)
	assert.Equal(1, b[0].Priority)
	assert.Equal(2, b[1].Priority)
}

func TestBrokerValidationHost(t *testing.T) {
	assert := assert.New(t)

	// invalid host, too long
	configStr := `
[[broker."sango/2"]]
    host = "192.168.1.22aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
    port = 1883
`
	conf, err := config.LoadConfigByte([]byte(configStr))
	_, err = NewBrokers(conf, make(chan message.Message))
	assert.NotNil(err)
}

func TestBrokerValidationPort(t *testing.T) {
	assert := assert.New(t)
	configStr := `
[[broker."sango/2"]]
    host = "192.168.1.22"
    port = 65536
`
	conf, err := config.LoadConfigByte([]byte(configStr))
	_, err = NewBrokers(conf, make(chan message.Message))
	assert.NotNil(err)
}

func TestBrokerValidationPriority(t *testing.T) {
	assert := assert.New(t)
	configStr := `
	[[broker."sango/10"]]
    host = "192.168.1.22"
    port = 1883
`
	conf, err := config.LoadConfigByte([]byte(configStr))
	_, err = NewBrokers(conf, make(chan message.Message))
	assert.NotNil(err)

	configStr = `
	[[broker."sango/0"]]
    host = "192.168.1.22"
    port = 1883
`
	conf, err = config.LoadConfigByte([]byte(configStr))
	_, err = NewBrokers(conf, make(chan message.Message))
	assert.NotNil(err)

}

func TestBrokerValidationWill(t *testing.T) {
	assert := assert.New(t)
	configStr := `
	[[broker."sango/1"]]
    host = "192.168.1.22"
    port = 1883
    will_message = "will"
`
	conf, err := config.LoadConfigByte([]byte(configStr))
	b, err := NewBrokers(conf, make(chan message.Message))
	assert.Nil(err)
	assert.Equal(1, len(b))
	assert.Equal([]byte("will"), b[0].WillMessage)

	configStr = `
	[[broker."sango/1"]]
    host = "192.168.1.22"
    port = 1883
    will_message = "\\x01\\x0f"
`
	conf, err = config.LoadConfigByte([]byte(configStr))
	b, err = NewBrokers(conf, make(chan message.Message))
	assert.Nil(err)
	assert.Equal(1, len(b))
	assert.Equal([]byte{1, 15}, b[0].WillMessage)

	// either will message has invalid binary, not error, just warn
	configStr = `
	[[broker."sango/1"]]
    host = "192.168.1.22"
    port = 1883
    will_message = "\\x01\\x0fffff"
`
	conf, err = config.LoadConfigByte([]byte(configStr))
	b, err = NewBrokers(conf, make(chan message.Message))
	assert.Nil(err)
	assert.Equal(1, len(b))
	assert.Equal([]byte{1, 15}, b[0].WillMessage)
}

func TestBrokerValidationTls(t *testing.T) {
	assert := assert.New(t)

	// check broker validation
	configStr := `
	[[broker."sango/1"]]
   host = "localhost"
   port = 8883
   tls = true
   cacert = "../tests/mosquitto/ca.pem"
`
	conf, err := config.LoadConfigByte([]byte(configStr))
	b, err := NewBrokers(conf, make(chan message.Message))
	assert.Nil(err)
	assert.Equal(1, len(b))

	// check broker validation fail if cacert is missing
	configStr = `
	[[broker."sango/1"]]
    host = "localhost"
    port = 8883
    tls = true
`
	conf, err = config.LoadConfigByte([]byte(configStr))
	assert.Nil(err)
	b, err = NewBrokers(conf, make(chan message.Message))
	assert.NotNil(err)
}

func TestGenerateTopic(t *testing.T) {
	assert := assert.New(t)
	b := &Broker{
		GatewayName: "gw",
		Name:        "b",
		TopicPrefix: "prefix",
	}

	msg1 := &message.Message{
		Sender: "s",
		Type:   "t",
	}
	t1, err := b.GenerateTopic(msg1)
	assert.Nil(err)
	assert.Equal("prefix/gw/s/t/publish", t1.Str)

	msg2 := &message.Message{
		Sender: "s1",
	}
	t2, err := b.GenerateTopic(msg2)
	assert.Nil(err)
	assert.Equal("prefix/gw/s1//publish", t2.Str)
}

func TestGenerateTopicStatus(t *testing.T) {
	assert := assert.New(t)
	b := &Broker{
		GatewayName: "gw",
		Name:        "b",
		TopicPrefix: "prefix",
	}

	msg1 := &message.Message{
		Sender: "status",
		Type:   "t",
	}
	t1, err := b.GenerateTopic(msg1)
	assert.Nil(err)
	assert.Equal("prefix/", t1.Str)

	msg2 := &message.Message{
		Topic:  "$SYS/gw/cpu/total",
		Sender: "status",
		Type:   "t",
	}
	t2, err := b.GenerateTopic(msg2)
	assert.Nil(err)
	assert.Equal("prefix/$SYS/gw/cpu/total", t2.Str)
}

func TestBrokersPrioritySort(t *testing.T) {
	assert := assert.New(t)

	// Broker priority range is 1-3.
	b3 := &Broker{
		GatewayName: "gw1",
		Name:        "b",
		Priority:    3,
	}
	b1 := &Broker{
		GatewayName: "gw1",
		Name:        "b",
		Priority:    1,
	}
	b2 := &Broker{
		GatewayName: "gw1",
		Name:        "b",
		Priority:    2,
	}
	var bs Brokers
	bs = append(bs, b3, b1, b2)
	assert.Equal(3, bs[0].Priority)
	assert.Equal(1, bs[1].Priority)
	assert.Equal(2, bs[2].Priority)

	sort.Sort(bs)
	assert.Equal(1, bs[0].Priority)
	assert.Equal(2, bs[1].Priority)
	assert.Equal(3, bs[2].Priority)

}
