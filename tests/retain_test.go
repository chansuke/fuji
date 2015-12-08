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
	"testing"
	"time"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/stretchr/testify/assert"

	"github.com/shiguredo/fuji"
	"github.com/shiguredo/fuji/broker"
	"github.com/shiguredo/fuji/config"
	"github.com/shiguredo/fuji/device"
	"github.com/shiguredo/fuji/gateway"
)

// TestRetainJustPublish tests
// 1. connect gateway to local broker
// 2. send data with retaind flag dummy device normaly
func TestRetainJustPublish(t *testing.T) {
	assert := assert.New(t)

	configStr := `
	[gateway]
	
	    name = "retainham"
	
	[[broker."local/1"]]
	
	    host = "localhost"
	    port = 1883
	
	[device."doraretain"]
	
	    type = "dummy"
	    broker = "local"
	    qos = 0
	
	    interval = 10
	    payload = "Hello world retain true."
	    retain = true
`

	conf, err := config.LoadConfigByte([]byte(configStr))
	assert.Nil(err)

	commandChannel := make(chan string)
	go fuji.StartByFileWithChannel(conf, commandChannel)

	time.Sleep(2 * time.Second)
}

// TestRetainSubscribePublishClose
// 1. connect gateway to local broker
// 2. send data with retaind flag from dummy device
// 3. disconnect
// 4. reconnect
// 5. subscirbe and receive data
func TestRetainSubscribePublishClose(t *testing.T) {
	assert := assert.New(t)
	configStr := `
	[gateway]
	
	    name = "testRetainafterclose"
	
	[[broker."local/1"]]
	
	    host = "localhost"
	    port = 1883
	
	[device."dora"]
	
	    type = "dummy"
	    broker = "local"
	    qos = 0
	
	    interval = 10
	    payload = "Hello retained world to subscriber after close."
	
	    retain = true
`
	conf, err := config.LoadConfigByte([]byte(configStr))
	assert.Nil(err)
	isRetain := true
	generalPubSubTest(t, conf, isRetain)
}

// TestNoRetainSubscribePublishClose
// 1. connect gateway to local broker
// 2. send data without retaind flag from dummy device
// 3. disconnect
// 4. reconnect
// 5. subscirbe and receive data
func TestNoRetainSubscribePublishClose(t *testing.T) {
	assert := assert.New(t)
	configStr := `
	[gateway]
	
	    name = "testNoRetainafterclose"
	
	[[broker."local/1"]]
	
	    host = "localhost"
	    port = 1883
	
	[device."dora"]
	
	    type = "dummy"
	    broker = "local"
	    qos = 0
	
	    interval = 10
	    payload = "Hello retained world to subscriber after close."
`
	conf, err := config.LoadConfigByte([]byte(configStr))
	assert.Nil(err)
	isRetain := false
	generalPubSubTest(t, conf, isRetain)
}

func generalPubSubTest(t *testing.T, conf config.Config, isRetain bool) {
	assert := assert.New(t)

	commandChannel := make(chan string)
	go fuji.StartByFileWithChannel(conf, commandChannel)

	gw, err := gateway.NewGateway(conf)
	if err != nil {
		t.Error("Cannot make Gateway")
	}

	brokerList, err := broker.NewBrokers(conf, gw.BrokerChan)
	if err != nil {
		t.Error("Cannot make BrokerList")
	}

	devChan := device.NewDeviceChannel()
	gw.DeviceChannels = append(gw.DeviceChannels, devChan)
	dummyDevice, err := device.NewDummyDevice(conf.Sections[2], brokerList, devChan)
	if err != nil {
		t.Error("Cannot make DummyDeviceList")
	}

	go func() {
		time.Sleep(2 * time.Second)

		// kill publisher
		gw.Stop()

		time.Sleep(2 * time.Second)

		subscriberChannel, err := setupRetainSubscriber(gw, brokerList[0], &dummyDevice)
		if err != config.Error("") {
			t.Error(err)
		}

		// check Retained message
		select {
		case retainedMessage := <-subscriberChannel:
			if !isRetain {
				assert.Equal("retained message arrived", "no retain message shall come")
			}
			retainedTopic := retainedMessage[0]
			retainedPayload := retainedMessage[1]

			expectedTopic := fmt.Sprintf("%s/%s/%s/%s/publish", brokerList[0].TopicPrefix, gw.Name, dummyDevice.Name, dummyDevice.Type)
			expectedPayload := dummyDevice.Payload

			assert.Equal(expectedTopic, retainedTopic)
			assert.Equal(expectedPayload, retainedPayload)

		case <-time.After(time.Second * 2):
			if isRetain {
				assert.Equal("subscribe completed in 11 sec", "not completed")
			}
		}
	}()
	time.Sleep(5 * time.Second)
}

// setupRetainSubscriber returnes channel in order to read messages with retained flag
func setupRetainSubscriber(gw *gateway.Gateway, broker *broker.Broker, dummyDevice *device.DummyDevice) (chan [2]string, config.Error) {
	// Setup MQTT pub/sub client to confirm published content.
	//
	messageOutputChannel := make(chan [2]string)

	opts := MQTT.NewClientOptions()
	brokerUrl := fmt.Sprintf("tcp://%s:%d", broker.Host, broker.Port)
	opts.AddBroker(brokerUrl)
	opts.SetClientID(gw.Name + "testSubscriber") // to distinguish MQTT client from publisher
	opts.SetCleanSession(false)
	opts.SetDefaultPublishHandler(func(client *MQTT.Client, msg MQTT.Message) {
		messageOutputChannel <- [2]string{msg.Topic(), string(msg.Payload())}
	})

	client := MQTT.NewClient(opts)
	if client == nil {
		return nil, config.Error("NewClient failed")
	}

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, config.Error(fmt.Sprintf("NewClient Start failed %q", token.Error()))
	}
	qos := 0
	retainedTopic := fmt.Sprintf("%s/%s/%s/%s/publish", broker.TopicPrefix, gw.Name, dummyDevice.Name, dummyDevice.Type)
	client.Subscribe(retainedTopic, byte(qos), func(client *MQTT.Client, msg MQTT.Message) {
	})

	return messageOutputChannel, config.Error("")
}
