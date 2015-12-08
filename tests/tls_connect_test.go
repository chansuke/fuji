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

var configStr = `
[gateway]

    name = "hamlocalconnect"

[[broker."mosquitto/1"]]

    host = "localhost"
    port = 8883
    tls = true
    cacert = "mosquitto/ca.pem"

    retry_interval = 10


[device."dora"]
    type = "dummy"

    broker = "mosquitto"
    qos = 0

    interval = 10
    payload = "connect local pub only Hello world."
`

// TestTLSConnectLocalPub
func TestTLSConnectLocalPub(t *testing.T) {
	assert := assert.New(t)

	conf, err := config.LoadConfigByte([]byte(configStr))
	assert.Nil(err)
	commandChannel := make(chan string)
	go fuji.StartByFileWithChannel(conf, commandChannel)
	time.Sleep(2 * time.Second)

}

// TestTLSConnectLocalPubSub
// 1. connect gateway to local broker with TLS
// 2. send data from dummy
// 3. check subscribe
func TestTLSConnectLocalPubSub(t *testing.T) {
	assert := assert.New(t)

	// pub/sub test to broker on localhost
	// dummydevice is used as a source of published message
	// publised messages confirmed by subscriber

	// get config
	conf, err := config.LoadConfigByte([]byte(configStr))
	assert.Nil(err)

	// get Gateway
	gw, err := gateway.NewGateway(conf)
	assert.Nil(err)

	// get Broker
	brokerList, err := broker.NewBrokers(conf, gw.BrokerChan)
	assert.Nil(err)

	// get DummyDevice
	dummyDevice, err := device.NewDummyDevice(conf.Sections[2], brokerList, device.NewDeviceChannel())
	assert.Nil(err)
	assert.NotNil(dummyDevice)

	// Setup MQTT pub/sub client to confirm published content.
	//
	subscriberChannel := make(chan [2]string)

	opts := MQTT.NewClientOptions()
	url := fmt.Sprintf("ssl://%s:%d", brokerList[0].Host, brokerList[0].Port)
	opts.AddBroker(url)
	opts.SetClientID(fmt.Sprintf("prefix%s", gw.Name))
	opts.SetCleanSession(false)

	tlsConfig, err := broker.NewTLSConfig(brokerList[0])
	assert.Nil(err)
	opts.SetTLSConfig(tlsConfig)

	client := MQTT.NewClient(opts)
	assert.Nil(err)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		assert.Nil(token.Error())
		fmt.Println(token.Error())
	}

	qos := 0
	expectedTopic := fmt.Sprintf("/%s/%s/%s/publish", gw.Name, dummyDevice.Name, dummyDevice.Type)
	expectedMessage := fmt.Sprintf("%s", dummyDevice.Payload)
	fmt.Printf("expetcted topic: %s\nexpected message%s", expectedTopic, expectedMessage)
	client.Subscribe(expectedTopic, byte(qos), func(client *MQTT.Client, msg MQTT.Message) {
		subscriberChannel <- [2]string{msg.Topic(), string(msg.Payload())}
	})

	// wait for 1 publication of dummy worker
	select {
	case message := <-subscriberChannel:
		assert.Equal(expectedTopic, message[0])
		assert.Equal(expectedMessage, message[1])
	case <-time.After(time.Second * 11):
		assert.Equal("subscribe completed in 11 sec", "not completed")
	}

	client.Disconnect(20)
}
