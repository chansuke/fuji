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

package fuji

import (
	log "github.com/Sirupsen/logrus"

	"github.com/shiguredo/fuji/broker"
	"github.com/shiguredo/fuji/config"
	"github.com/shiguredo/fuji/device"
	"github.com/shiguredo/fuji/gateway"
)

// Start make command channel and start gateway.
func Start(configPath string) {
	conf, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("loading config file faild, %v", err)
	}

	commandChannel := make(chan string)

	err = StartByFileWithChannel(conf, commandChannel)
	if err != nil {
		log.Error(err)
	}
}

// StartByFileWithChannel starts Gateway with command Channel
func StartByFileWithChannel(conf config.Config, commandChannel chan string) error {
	gw, err := gateway.NewGateway(conf)
	if err != nil {
		log.Fatalf("gateway create error, %v", err)
	}
	brokerList, err := broker.NewBrokers(conf, gw.BrokerChan)
	if err != nil {
		log.Fatalf("broker(s) create error, %v", err)
	}
	deviceList, deviceChannels, err := device.NewDevices(conf, brokerList)
	if err != nil {
		log.Fatalf("device create error, %v", err)
	}

	gw.Devices = deviceList
	gw.Brokers = brokerList
	gw.CmdChan = commandChannel
	gw.DeviceChannels = deviceChannels

	status, err := device.NewStatus(conf)
	if err != nil {
		log.Warnf("status create error, %v", err)
		// run whenever status created
	} else {
		gw.Devices = append(gw.Devices, status)
	}

	// add to brokers subscribed
	for _, device := range gw.Devices {
		err := device.AddSubscribe()
		if err != nil {
			log.Errorf("device subscribe error, %v", err)
			continue
		}
	}

	// Start brokers and devices
	for _, b := range gw.Brokers {
		err := b.MQTTClientSetup(gw.Name)
		if err != nil {
			log.Errorf("MQTTClientSetup failed, %v", err)
			continue
		}
	}
	for _, device := range gw.Devices {
		err := device.Start(gw.MsgChan)
		if err != nil {
			log.Errorf("device start error, %v", err)
			continue
		}
	}

	// start gateway
	return gw.Start()
}
