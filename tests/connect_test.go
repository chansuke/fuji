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
	"github.com/shiguredo/fuji/device"
	"github.com/shiguredo/fuji/gateway"
	"github.com/shiguredo/fuji/toml"
)

// publish test to broker on localhost
// dummydevice is used as a source of published message
func TestConnectLocalPub(t *testing.T) {

	go fuji.Start("connectlocalpub.ini")

	time.Sleep(2 * time.Second)
}

// TestConnectLocalPubSub tests
// 1. connect gateway to local broker
// 2. send data from dummy
// 3. check subscribe
func TestConnectLocalPubSub(t *testing.T) {
	assert := assert.New(t)

	// pub/sub test to broker on localhost
	// dummydevice is used as a source of published message
	// publised messages confirmed by subscriber

	// get config
	conf, err := toml.LoadConfig("connectlocalpubsub.toml")
	assert.Nil(err)

	// get Gateway
	gw, err := gateway.NewGateway(conf)
	assert.Nil(err)

	// get Broker
	brokerList, err := broker.NewBrokers(conf, gw.BrokerChan)
	assert.Nil(err)

	// get DummyDevice
	dummyDevice, err := device.NewDummyDevice(conf.Sections[3], brokerList, device.NewDeviceChannel())
	assert.Nil(err)
	assert.NotNil(dummyDevice)

	// Setup MQTT pub/sub client to confirm published content.
	//
	subscriberChannel := make(chan [2]string)

	opts := MQTT.NewClientOptions()
	url := fmt.Sprintf("tcp://%s:%d", brokerList[0].Host, brokerList[0].Port)
	opts.AddBroker(url)
	opts.SetClientID(gw.Name)
	opts.SetCleanSession(false)
	opts.SetDefaultPublishHandler(func(client *MQTT.Client, msg MQTT.Message) {
		subscriberChannel <- [2]string{msg.Topic(), string(msg.Payload())}
	})

	client := MQTT.NewClient(opts)
	assert.Nil(err)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		assert.Nil(token.Error())
	}

	qos := 0
	topic := "#"
	client.Subscribe(topic, byte(qos), func(client *MQTT.Client, msg MQTT.Message) {
	})

	// TODO: should be write later
	/*
		channel := fuji.SetupMQTTChannel(client, gateway, brokerList[0])

		// Setup DummyDevice to publish test payload

		dummyDevice.Start(channel)

		// wait for 1 publication of dummy worker
		message := <-subscriberChannel
		assert.Equal("dummy", message)

		client.Disconnect(250)
	*/
}
