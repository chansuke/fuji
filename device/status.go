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
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	validator "gopkg.in/validator.v2"

	"github.com/shiguredo/fuji/config"
	"github.com/shiguredo/fuji/message"
)

type CPUStatus struct {
	GatewayName string
	BrokerName  string
	CpuTimes    []string
}
type MemoryStatus struct {
	GatewayName   string
	BrokerName    string
	VirtualMemory []string
}
type IpAddressStatus struct {
	GatewayName string
	BrokerName  string
	Interfaces  []string
}
type Status struct {
	Name        string `validate:"max=256,regexp=[^/]+,validtopic"`
	GatewayName string
	BrokerName  string
	Interval    int
	CPU         CPUStatus
	Memory      MemoryStatus
	IpAddress   IpAddressStatus
}

func (device Status) String() string {
	return fmt.Sprintf("%#v", device)
}

// Get returns CPU status by Message.
// Note: currently, only one CPU status is retrieved.
func (c CPUStatus) Get() []message.Message {
	ret := []message.Message{}

	cpuTimes, err := cpu.CPUTimes(false)
	if err == nil {
		for _, t := range c.CpuTimes {
			msg := message.Message{
				Sender:     "status",
				Type:       "status",
				BrokerName: c.BrokerName,
			}
			var body string
			switch t {
			case "user":
				body = strconv.Itoa(int(cpuTimes[0].User))
			case "system":
				body = strconv.Itoa(int(cpuTimes[0].System))
			case "idle":
				body = strconv.Itoa(int(cpuTimes[0].Idle))
			case "nice":
				body = strconv.Itoa(int(cpuTimes[0].Nice))
			case "iowait":
				body = strconv.Itoa(int(cpuTimes[0].Iowait))
			case "irq":
				body = strconv.Itoa(int(cpuTimes[0].Irq))
			case "softirq":
				body = strconv.Itoa(int(cpuTimes[0].Softirq))
			case "guest":
				body = strconv.Itoa(int(cpuTimes[0].Guest))
			}

			msg.Body = []byte(body)
			topic, err := genTopic(c.GatewayName, "cpu", "cpu_times", t)
			if err != nil {
				log.Errorf("invalid topic, %s/%s/%s/%s", c.GatewayName, "cpu", "cpu_times", t)
				continue
			}
			msg.Topic = topic

			ret = append(ret, msg)
		}

	} else {
		log.Warnf("cpu get err, %v", err)
	}

	return ret
}
func (m MemoryStatus) Get() []message.Message {
	ret := []message.Message{}

	vmem, err := mem.VirtualMemory()
	if err == nil {
		for _, t := range m.VirtualMemory {
			msg := message.Message{
				Sender:     "status",
				Type:       "status",
				BrokerName: m.BrokerName,
			}
			var body string
			switch t {
			case "total":
				body = strconv.Itoa(int(vmem.Total))
			case "available":
				body = strconv.Itoa(int(vmem.Available))
			case "percent":
				body = fmt.Sprintf("%v", vmem.UsedPercent)
			case "used":
				body = strconv.Itoa(int(vmem.Used))
			case "free":
				body = strconv.Itoa(int(vmem.Free))
			}
			msg.Body = []byte(body)
			topic, err := genTopic(m.GatewayName, "memory", "virtual_memory", t)
			if err != nil {
				log.Errorf("invalid topic, %s/%s/%s/%s", m.GatewayName, "memory", "virtual_memory", t)
				continue
			}
			msg.Topic = topic

			ret = append(ret, msg)
		}
	} else {
		log.Warnf("virtual_memory get error, %v", err)
	}
	return ret
}

func (i IpAddressStatus) Get() []message.Message {
	ret := []message.Message{}

	addresses, err := net.Interfaces()
	if err == nil {
		msg := message.Message{
			Sender:     "status",
			Type:       "status",
			BrokerName: i.BrokerName,
		}
		body, err := json.Marshal(addresses)
		if err != nil {
			log.Errorf("json encode error %s", err)
		}
		msg.Body = []byte(body)

		topic, err := genTopic(i.GatewayName, "ip_address", "interface", "all")
		if err != nil {
			log.Errorf("invalid topic, %s/%s/%s/%s", i.GatewayName, "ip_address", "interface", "all")
		}
		msg.Topic = topic

		ret = append(ret, msg)
	} else {
		log.Warnf("ip_address get err, %v", err)
	}
	return ret
}

// NewStatus returnes status from config file, not config.Sections.
func NewStatus(conf config.Config) (Devicer, error) {
	ret := Status{
		Name:        "status",
		GatewayName: conf.GatewayName,
	}

	// first, search "status" section
	for _, section := range conf.Sections {
		if section.Type != "status" {
			continue
		}

		if section.Name != "" { // skip if status child group
			continue
		}
		bname, ok := section.Values["broker"]
		if !ok {
			return ret, fmt.Errorf("status does not have broker name")
		}

		for _, b := range conf.BrokerNames {
			if b == bname {
				ret.BrokerName = b
			}
		}
		if ret.BrokerName == "" {
			return ret, fmt.Errorf("broker does not exists: %s", section.Name)
		}
		interval, err := strconv.Atoi(section.Values["interval"])
		if err != nil {
			return ret, err
		} else {
			ret.Interval = int(interval)
		}
	}

	// status-wide settings done. now walk to childs
	for _, section := range conf.Sections {
		if section.Type != "status" || section.Name == "" {
			continue
		}
		switch section.Name {
		case "cpu":
			cpu_times := parseStatus(section.Values["cpu_times"])

			cpu := CPUStatus{
				GatewayName: conf.GatewayName,
				BrokerName:  ret.BrokerName,
			}
			if len(cpu_times) > 0 {
				cpu.CpuTimes = cpu_times
			}

			ret.CPU = cpu
		case "memory":
			virtual_memory := parseStatus(section.Values["virtual_memory"])

			mem := MemoryStatus{
				GatewayName: conf.GatewayName,
				BrokerName:  ret.BrokerName,
			}
			if len(virtual_memory) > 0 {
				mem.VirtualMemory = virtual_memory
			}
			ret.Memory = mem

		case "ip_address":
			interfaces := parseStatus(section.Values["ip_address"])

			ip_address := IpAddressStatus{
				GatewayName: conf.GatewayName,
				BrokerName:  ret.BrokerName,
			}
			if len(interfaces) > 0 {
				ip_address.Interfaces = interfaces
			}
			ret.IpAddress = ip_address
		default:
			log.Errorf("unknown status type: %v", section.Name)
			continue
		}

	}

	if err := ret.Validate(); err != nil {
		return ret, err
	}

	if ret.Interval == 0 {
		return ret, fmt.Errorf("no status found")
	}

	return ret, nil
}

func (device *Status) Validate() error {
	validator := validator.NewValidator()
	validator.SetValidationFunc("validtopic", config.ValidMqttPublishTopic)
	if err := validator.Validate(device); err != nil {
		return err
	}
	return nil
}

func (device Status) Start(channel chan message.Message) error {
	log.Infof("start status")
	go func() {
		for {
			msgs := make([]message.Message, 0, 10)

			msgs = append(msgs, device.CPU.Get()...)
			msgs = append(msgs, device.Memory.Get()...)
			msgs = append(msgs, device.IpAddress.Get()...)
			if len(msgs) > 0 {
				for _, msg := range msgs {
					channel <- msg
				}
			}

			time.Sleep(time.Duration(device.Interval) * time.Second)
		}
	}()
	return nil
}

func (device Status) Stop() error {
	log.Infof("closing status: %v", device.Name)
	return nil
}

func (device Status) DeviceType() string {
	return "status"
}

// parseStatus parse fields in the status childs
// ex: user, system, idle, nice, => []string{"user", "system", "idle", "nice"}
func parseStatus(buf string) []string {
	ret := []string{}

	for _, r := range strings.Split(buf, ",") {
		if r == "" {
			continue
		}
		ret = append(ret, strings.TrimSpace(r))
	}

	return ret
}

func genTopic(gwName, main, sub, item string) (string, error) {
	topic := fmt.Sprintf("$SYS/gateway/%s/%s/%s/%s", gwName, main, sub, item)
	return topic, nil
}

func (device Status) AddSubscribe() error {
	// Status does not subscibe
	return nil
}
