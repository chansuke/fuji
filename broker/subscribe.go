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
	"fmt"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
)

type Subscribed struct {
	sync.Mutex

	list map[string]byte
}

func NewSubscribed() Subscribed {
	s := Subscribed{
		list: make(map[string]byte),
	}
	return s
}
func (s Subscribed) Length() int {
	return len(s.list)
}
func (s Subscribed) List() map[string]byte {
	return s.list
}

func (s Subscribed) Add(topic string, qos byte) error {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.list[topic]; ok {
		log.Warnf("topic %v is already added subscribed, override it", topic)
	}
	// override if already exists
	s.list[topic] = qos

	return nil
}

func (s Subscribed) Delete(topic string) error {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.list[topic]; !ok {
		return fmt.Errorf("topic %v is not in the subscribed", topic)
	}
	delete(s.list, topic)
	return nil
}
func (b *Broker) AddSubscribed(deviceName string, qos byte) error {
	t := strings.Join([]string{b.TopicPrefix, b.GatewayName, deviceName, "subscribe"}, "/")
	log.Infof("subscribe: %#v", t)
	return b.Subscribed.Add(t, qos)
}
func (b *Broker) DeleteSubscribed(deviceName string, qos byte) error {
	t := strings.Join([]string{b.TopicPrefix, b.GatewayName, deviceName}, "/")
	return b.Subscribed.Delete(t)
}
