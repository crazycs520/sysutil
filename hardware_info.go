package sysutil

import (
	"strconv"

	pb "github.com/pingcap/kvproto/pkg/diagnosticspb"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/net"
)

var singleDevicesHardwareInfoFns = []struct {
	tp   string
	name string
	fn   func() (interface{}, error)
}{
	{"host", "host", getHost},
	{"mem", "virtual", getVirtualMemStat},
}

var multiDevicesHardInfoInfoFns = []struct {
	tp string
	fn func() (map[string]interface{}, error)
}{
	{"cpu", getCPU},
	{"net", getNet},
	{"disk", getDisk},
}

func getHardwareInfo() []*pb.ServerInfoItem {
	items := make([]*pb.ServerInfoItem, 0, len(singleDevicesLoadInfoFns))
	for _, f := range singleDevicesHardwareInfoFns {
		data, err := f.fn()
		if err != nil {
			continue
		}
		item, err := convertToServerInfoItems(f.tp, f.name, data)
		if err != nil {
			continue
		}
		items = append(items, item)
	}
	for _, f := range multiDevicesHardInfoInfoFns {
		ds, err := f.fn()
		if err != nil {
			continue
		}

		for k, data := range ds {
			item, err := convertToServerInfoItems(f.tp, k, data)
			if err != nil {
				continue
			}
			items = append(items, item)
		}
	}
	return items
}

func getCPU() (map[string]interface{}, error) {
	cpus, err := cpu.Info()
	if err != nil {
		return nil, err
	}
	m := make(map[string]interface{}, len(cpus))
	for _, c := range cpus {
		name := "cpu" + strconv.FormatInt(int64(c.CPU), 10)
		m[name] = c
	}
	return m, nil
}

func getDisk() (map[string]interface{}, error) {
	parts, err := disk.Partitions(true)
	if err != nil {
		return nil, err
	}
	m := make(map[string]interface{}, len(parts))
	for _, part := range parts {
		m[part.Device] = part
	}
	return m, nil
}

func getNet() (map[string]interface{}, error) {
	nets, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	m := make(map[string]interface{}, len(nets))
	for _, n := range nets {
		m[n.Name] = n
	}
	return m, nil
}

func getHost() (interface{}, error) {
	return host.Info()
}
