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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/stretchr/testify/assert"

	"github.com/shiguredo/fuji"
	"github.com/shiguredo/fuji/broker"
	"github.com/shiguredo/fuji/config"
	"github.com/shiguredo/fuji/gateway"
)

var tmpTomlName = ".tmp.toml"

// TestWillJustPublish tests
// 1. connect localhost broker with will message
// 2. send data from a dummy device
// 3. disconnect
func TestWillJustPublish(t *testing.T) {
	assert := assert.New(t)

	configStr := `
	[gateway]
	    name = "willjustpublishham"
	[[broker."local/1"]]
	    host = "localhost"
	    port = 1883
	    will_message = "no letter is good letter."
	[device."dora"]
	    type = "dummy"
	    broker = "local"
	    qos = 0
	    interval = 10
	    payload = "Hello will just publish world."
`
	conf, err := config.LoadConfigByte([]byte(configStr))
	assert.Nil(err)
	commandChannel := make(chan string)
	go fuji.StartByFileWithChannel(conf, commandChannel)
	time.Sleep(5 * time.Second)

	//	fuji.Stop()
}

// TestWillWithPrefixSubscribePublishClose
// 1. connect subscriber and publisher to localhost broker with will message with prefixed topic
// 2. send data from a dummy device
// 3. force disconnect
// 4. check subscriber does not receives will message immediately
func TestWillWithPrefixSubscribePublishClose(t *testing.T) {
	assert := assert.New(t)

	configStr := `
	[gateway]
	    name = "testprefixwill"
	[[broker."local/1"]]
	    host = "localhost"
	    port = 1883
	    will_message = "no letter is good letter."
	    topic_prefix = "prefix"
	[device."dora"]
	    type = "dummy"
	    broker = "local"
	    qos = 0
	    interval = 10
	    payload = "Hello will with prefix."
`
	expectedWill := true
	ok := genericWillTestDriver(t, configStr, "prefix/testprefixwill/will", []byte("no letter is good letter."), expectedWill)
	assert.True(ok, "Failed to receive Will with prefix message")
}

// TestNoWillSubscribePublishClose
// 1. connect subscriber and publisher to localhost broker without will message
// 2. send data from a dummy device
// 3. force disconnect
// 4. check subscriber does not receives will message immediately
func TestNoWillSubscribePublishClose(t *testing.T) {
	assert := assert.New(t)

	configStr := `
	[gateway]
	    name = "testnowillafterclose"
	[[broker."local/1"]]
	    host = "localhost"
	    port = 1883
	[device."dora"]
	    type = "dummy"
	    broker = "local"
	    qos = 0
	    interval = 10
	    payload = "Hello will just publish world."
`
	expectedWill := false
	ok := genericWillTestDriver(t, configStr, "/testnowillafterclose/will", []byte(""), expectedWill)
	assert.False(ok, "Failed to receive Will message")
}

// TestWillSubscribePublishClose
// 1. connect subscriber and publisher to localhost broker with will message
// 2. send data from a dummy device
// 3. force disconnect
// 4. check subscriber receives will message
func TestWillSubscribePublishClose(t *testing.T) {
	assert := assert.New(t)

	configStr := `
	[gateway]
	    name = "testwillafterclose"
	[[broker."local/1"]]
	    host = "localhost"
	    port = 1883
	    will_message = "good letter is no letter."
	[device."dora"]
	    type = "dummy"
	    broker = "local"
	    qos = 0
	    interval = 10
	    payload = "Hello will just publish world."
`
	expectedWill := true
	ok := genericWillTestDriver(t, configStr, "/testwillafterclose/will", []byte("good letter is no letter."), expectedWill)
	assert.True(ok, "Failed to receive Will message")
}

// TestWillSubscribePublishCloseEmpty
// 1. connect subscriber and publisher to localhost broker with will message
// 2. send data from a dummy device
// 3. force disconnect
// 4. check subscriber receives will message
func TestWillSubscribePublishCloseEmpty(t *testing.T) {
	configStr := `
	[gateway]
	    name = "testwillaftercloseemptywill"
	[[broker."local/1"]]
	    host = "localhost"
	    port = 1883
	    will_message = ""
	[device."dora"]
	    type = "dummy"
	    broker = "local"
	    qos = 0
	    interval = 10
	    payload = "Hello will just publish world."
`
	expectedWill := true
	ok := genericWillTestDriver(t, configStr, "/testwillaftercloseemptywill/will", []byte{}, expectedWill)
	if !ok {
		t.Error("Failed to receive Empty Will message")
	}
}

func TestWillSubscribePublishBinaryWill(t *testing.T) {
	configStr := `
	[gateway]
	    name = "binary"
	[[broker."local/1"]]
	    host = "localhost"
	    port = 1883
	    will_message = "\\x01\\x02"
	[device."dora"]
	    type = "dummy"
	    broker = "local"
	    qos = 0
	    interval = 10
	    payload = "Hello will just publish world."
`
	expectedWill := true
	ok := genericWillTestDriver(t, configStr, "/binary/will", []byte{1, 2}, expectedWill)
	if !ok {
		t.Error("Failed to receive Empty Will message")
	}
}

func TestWillSubscribePublishWillWithWillTopic(t *testing.T) {
	configStr := `
	[gateway]
	    name = "with"
	[[broker."local/1"]]
	    host = "localhost"
	    port = 1883
	    will_message = "msg"
	    will_topic = "willtopic"
`
	expectedWill := true
	ok := genericWillTestDriver(t, configStr, "/willtopic", []byte("msg"), expectedWill)
	if !ok {
		t.Error("Failed to receive Empty Will message")
	}
}

func TestWillSubscribePublishWillWithNestedWillTopic(t *testing.T) {
	configStr := `
	[gateway]
	    name = "withnested"
	[[broker."local/1"]]
	    host = "localhost"
	    port = 1883
	    will_message = "msg"
	    will_topic = "willtopic/nested"
`
	expectedWill := true
	ok := genericWillTestDriver(t, configStr, "/willtopic/nested", []byte("msg"), expectedWill)
	if !ok {
		t.Error("Failed to receive nested willtopic Will message")
	}
}

// genericWillTestDriver
// 1. read config string
// 2. connect subscriber and publisher to localhost broker with will message
// 3. send data from a dummy device
// 4. force disconnect
// 5. check subscriber receives will message

func genericWillTestDriver(t *testing.T, configStr string, expectedTopic string, expectedPayload []byte, expectedWill bool) (ok bool) {
	assert := assert.New(t)

	conf, err := config.LoadConfigByte([]byte(configStr))
	assert.Nil(err)

	// write config string to temporal file
	f, err := os.Create(tmpTomlName)
	if err != nil {
		t.Error(err)
	}
	_, err = f.WriteString(configStr)
	if err != nil {
		t.Error(err)
	}
	f.Sync()

	// execute fuji as external process
	fujiPath, err := filepath.Abs("../fuji")
	if err != nil {
		t.Error("file path not found")
	}
	cmd := exec.Command(fujiPath, "-c", tmpTomlName)
	err = cmd.Start()
	if err != nil {
		t.Error(err)
	}

	// subscriber setup
	gw, err := gateway.NewGateway(conf)
	assert.Nil(err)

	brokers, err := broker.NewBrokers(conf, gw.BrokerChan)
	assert.Nil(err)

	subscriberChannel, err := setupWillSubscriber(gw, brokers[0])
	if err != config.Error("") {
		t.Error(err)
	}

	fin := make(chan bool)

	go func() {
		// check will message
		willCame := true
		select {
		case willMsg := <-subscriberChannel:
			if expectedWill {
				assert.Equal(expectedTopic, willMsg.Topic())
				assert.Equal(expectedPayload, willMsg.Payload())
				assert.Equal(byte(0), willMsg.Qos())
			} else {
				assert.Equal("NO will message received within 1 sec", "unexpected will message received.")
			}
		case <-time.After(time.Second * 2):
			if expectedWill {
				assert.Equal("will message received within 1 sec", "not completed")
			}
			willCame = false
		}
		fin <- willCame
	}()

	// wait for startup of external command process
	time.Sleep(time.Second * 1)

	// kill publisher
	err = cmd.Process.Kill()
	if err != nil {
		t.Error(err)
	}
	fmt.Println("broker killed for getting will message")

	ok = <-fin
	return ok
}

// setupWillSubscriber start subscriber process and returnes a channel witch can receive will message.
func setupWillSubscriber(gw *gateway.Gateway, broker *broker.Broker) (chan MQTT.Message, config.Error) {
	// Setup MQTT pub/sub client to confirm published content.
	//
	messageOutputChannel := make(chan MQTT.Message)

	opts := MQTT.NewClientOptions()
	brokerUrl := fmt.Sprintf("tcp://%s:%d", broker.Host, broker.Port)
	opts.AddBroker(brokerUrl)
	opts.SetClientID(gw.Name + "testSubscriber") // to distinguish MQTT client from publisher
	opts.SetCleanSession(false)
	opts.SetDefaultPublishHandler(func(client *MQTT.Client, msg MQTT.Message) {
		messageOutputChannel <- msg
	})
	willQoS := 0
	willTopic := broker.WillTopic
	fmt.Printf("expected will_topic: %s", willTopic)

	client := MQTT.NewClient(opts)
	if client == nil {
		return nil, config.Error("NewClient failed")
	}
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, config.Error(fmt.Sprintf("NewClient Start failed %q", token.Error()))
	}

	client.Subscribe(willTopic, byte(willQoS), func(client *MQTT.Client, msg MQTT.Message) {
		messageOutputChannel <- msg
	})

	return messageOutputChannel, config.Error("")
}
