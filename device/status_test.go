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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/shiguredo/fuji/config"
)

func TestParseStatus(t *testing.T) {
	assert := assert.New(t)

	r := parseStatus("user,system,idle")
	assert.Equal([]string{"user", "system", "idle"}, r)

	// with spaces
	r = parseStatus("  user,   system,  idle  ")
	assert.Equal([]string{"user", "system", "idle"}, r)

	// empty
	r = parseStatus("")
	assert.Equal([]string{}, r)
}
func TestStatus(t *testing.T) {
	assert := assert.New(t)

	configStr := `
[[broker."sango"]]
  host = "192.168.1.20"
  port = 1033
[[status."cpu"]]
  cpu_times = "user, system, idle, nice, iowait, irq, softirq, guest"
[[status."memory"]]
  virtual_memory = "total, available, percent, used, free"
[status]
  broker = "sango"
  interval = 10
`
	conf, err := config.LoadConfigByte([]byte(configStr))
	assert.Nil(err)
	tt, err := NewStatus(conf)
	assert.Nil(err)
	st, ok := tt.(Status)
	assert.True(ok)

	assert.Equal(st.BrokerName, "sango")
	assert.Equal(st.Interval, 10)

	assert.Equal(8, len(st.CPU.CpuTimes))
	assert.Equal(5, len(st.Memory.VirtualMemory))

}

func TestCPUGet(t *testing.T) {
	assert := assert.New(t)

	c := CPUStatus{
		CpuTimes: []string{"user", "system", "idle"},
	}
	assert.NotNil(c)
}
func TestMemoryGet(t *testing.T) {
	assert := assert.New(t)

	c := MemoryStatus{
		VirtualMemory: []string{"total", "available", "used", "percent"},
	}
	msgs := c.Get()
	assert.Equal(4, len(msgs))
}

func TestIpAddressGet(t *testing.T) {
	assert := assert.New(t)

	i := IpAddressStatus{
		Interfaces: []string{"all"},
	}
	assert.NotNil(i)
	msgs := i.Get()
	assert.True(len(msgs) > 0)
}
