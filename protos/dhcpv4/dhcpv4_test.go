// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package dhcpv4

import (
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/stretchr/testify/assert"

	"github.com/njcx/libbeat_v7/beat"
	"github.com/njcx/libbeat_v7/common"
	"github.com/njcx/libbeat_v7/logp"
	"github.com/njcx/packetbeat7_dpdk/procs"
	"github.com/njcx/packetbeat7_dpdk/protos"
	"github.com/njcx/packetbeat7_dpdk/publish"
)

var _ protos.UDPPlugin = &dhcpv4Plugin{}

var (
	_ dhcpv4.Option = &TextOption{}
	_ dhcpv4.Option = &IPAddressOption{}
	_ dhcpv4.Option = &IPAddressesOption{}
)

// Application layer data from packetbeat/tests/system/pcaps/dhcp.pcap.
var (
	dhcpRequest = []byte{
		0x01, 0x01, 0x06, 0x00, 0x00, 0x00, 0x3d, 0x1e, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0b, 0x82, 0x01, 0xfc, 0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x63, 0x82, 0x53, 0x63,
		0x35, 0x01, 0x03, 0x3d, 0x07, 0x01, 0x00, 0x0b, 0x82, 0x01, 0xfc, 0x42, 0x32, 0x04, 0xc0, 0xa8, 0x00, 0x0a, 0x36, 0x04,
		0xc0, 0xa8, 0x00, 0x01, 0x37, 0x04, 0x01, 0x03, 0x06, 0x2a, 0xff, 0x00,
	}

	dhcpACK = []byte{
		0x02, 0x01, 0x06, 0x00, 0x00, 0x00, 0x3d, 0x1e, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xc0, 0xa8, 0x00, 0x0a,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0b, 0x82, 0x01, 0xfc, 0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x63, 0x82, 0x53, 0x63,
		0x35, 0x01, 0x05, 0x3a, 0x04, 0x00, 0x00, 0x07, 0x08, 0x3b, 0x04, 0x00, 0x00, 0x0c, 0x4e, 0x33, 0x04, 0x00, 0x00, 0x0e,
		0x10, 0x36, 0x04, 0xc0, 0xa8, 0x00, 0x01, 0x01, 0x04, 0xff, 0xff, 0xff, 0x00, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
)

func TestParseDHCPRequest(t *testing.T) {
	logp.TestingSetup()
	p, err := newPlugin(true, nil, procs.ProcessesWatcher{}, nil)
	if err != nil {
		t.Fatal(err)
	}

	ipTuple := common.NewIPPortTuple(4, net.IP{0, 0, 0, 0}, 68, net.IP{255, 255, 255, 255}, 67)
	pkt := &protos.Packet{
		Ts:      time.Now(),
		Tuple:   ipTuple,
		Payload: dhcpRequest,
	}

	expected := beat.Event{
		Timestamp: pkt.Ts,
		Fields: common.MapStr{
			"type":   "dhcpv4",
			"status": "OK",
			"source": common.MapStr{
				"ip":    "0.0.0.0",
				"port":  68,
				"bytes": 272,
			},
			"destination": common.MapStr{
				"ip":   "255.255.255.255",
				"port": 67,
			},
			"client": common.MapStr{
				"ip":    "0.0.0.0",
				"port":  68,
				"bytes": 272,
			},
			"server": common.MapStr{
				"ip":   "255.255.255.255",
				"port": 67,
			},
			"event": common.MapStr{
				"category": []string{"network_traffic", "network"},
				"type":     []string{"connection", "protocol"},
				"dataset":  "dhcpv4",
				"kind":     "event",
				"start":    pkt.Ts,
			},
			"network": common.MapStr{
				"type":         "ipv4",
				"direction":    "unknown",
				"transport":    "udp",
				"protocol":     "dhcpv4",
				"bytes":        272,
				"community_id": "1:t9O1j0qj71O4wJM7gnaHtgmfev8=",
			},
			"related": common.MapStr{
				"ip": []string{"0.0.0.0", "255.255.255.255"},
			},
			"dhcpv4": common.MapStr{
				"client_mac":     "00:0b:82:01:fc:42",
				"flags":          "unicast",
				"hardware_type":  "Ethernet",
				"hops":           0,
				"op_code":        "bootrequest",
				"seconds":        0,
				"transaction_id": "0x00003d1e",
				"option": common.MapStr{
					"message_type": "request",
					"parameter_request_list": []string{
						"Subnet Mask",
						"Router",
						"Domain Name Server",
						"NTP Servers",
					},
					"requested_ip_address": "192.168.0.10",
					"server_identifier":    "192.168.0.1",
				},
			},
		},
	}

	actual := p.parseDHCPv4(pkt)
	if assert.NotNil(t, actual) {
		publish.MarshalPacketbeatFields(actual, nil, nil)
		t.Logf("DHCP event: %+v", actual)
		assertEqual(t, expected, *actual)
	}
}

func TestParseDHCPACK(t *testing.T) {
	p, err := newPlugin(true, nil, procs.ProcessesWatcher{}, nil)
	if err != nil {
		t.Fatal(err)
	}

	ipTuple := common.NewIPPortTuple(4, net.IP{192, 168, 0, 1}, 67, net.IP{192, 168, 0, 10}, 68)
	pkt := &protos.Packet{
		Ts:      time.Now(),
		Tuple:   ipTuple,
		Payload: dhcpACK,
	}

	expected := beat.Event{
		Timestamp: pkt.Ts,
		Fields: common.MapStr{
			"type":   "dhcpv4",
			"status": "OK",
			"source": common.MapStr{
				"ip":    "192.168.0.1",
				"port":  67,
				"bytes": 300,
			},
			"destination": common.MapStr{
				"ip":   "192.168.0.10",
				"port": 68,
			},
			"client": common.MapStr{
				"ip":   "192.168.0.10",
				"port": 68,
			},
			"server": common.MapStr{
				"ip":    "192.168.0.1",
				"port":  67,
				"bytes": 300,
			},
			"event": common.MapStr{
				"category": []string{"network_traffic", "network"},
				"type":     []string{"connection", "protocol"},
				"dataset":  "dhcpv4",
				"kind":     "event",
				"start":    pkt.Ts,
			},
			"network": common.MapStr{
				"type":         "ipv4",
				"direction":    "unknown",
				"transport":    "udp",
				"protocol":     "dhcpv4",
				"bytes":        300,
				"community_id": "1:VbRSZnvQqvLiQRhYHLrdVI17sLQ=",
			},
			"related": common.MapStr{
				"ip": []string{"192.168.0.1", "192.168.0.10"},
			},
			"dhcpv4": common.MapStr{
				"assigned_ip":    "192.168.0.10",
				"client_mac":     "00:0b:82:01:fc:42",
				"flags":          "unicast",
				"hardware_type":  "Ethernet",
				"hops":           0,
				"op_code":        "bootreply",
				"seconds":        0,
				"transaction_id": "0x00003d1e",
				"option": common.MapStr{
					"ip_address_lease_time_sec": 3600,
					"message_type":              "ack",
					"rebinding_time_sec":        3150,
					"renewal_time_sec":          1800,
					"server_identifier":         "192.168.0.1",
					"subnet_mask":               "255.255.255.0",
				},
			},
		},
	}

	actual := p.parseDHCPv4(pkt)
	if assert.NotNil(t, actual) {
		publish.MarshalPacketbeatFields(actual, nil, nil)
		t.Logf("DHCP event: %+v", actual)
		assertEqual(t, expected, *actual)
	}
}

func assertEqual(t testing.TB, expected, actual beat.Event) {
	assert.EqualValues(t, normalizeEvent(t, expected), normalizeEvent(t, actual))
}

func normalizeEvent(t testing.TB, event beat.Event) interface{} {
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}

	var out interface{}
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}
	return out
}
