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

package beater

import (
	"github.com/njcx/gopacket_dpdk/layers"

	"github.com/njcx/packetbeat7_dpdk/config"
	"github.com/njcx/packetbeat7_dpdk/decoder"
	"github.com/njcx/packetbeat7_dpdk/flows"
	"github.com/njcx/packetbeat7_dpdk/procs"
	"github.com/njcx/packetbeat7_dpdk/protos"
	"github.com/njcx/packetbeat7_dpdk/protos/icmp"
	"github.com/njcx/packetbeat7_dpdk/protos/tcp"
	"github.com/njcx/packetbeat7_dpdk/protos/udp"
	"github.com/njcx/packetbeat7_dpdk/publish"
	"github.com/njcx/packetbeat7_dpdk/sniffer"
)

func workerFactory(publisher *publish.TransactionPublisher, protocols *protos.ProtocolsStruct, watcher procs.ProcessesWatcher, flows *flows.Flows, cfg config.Config) func(dl layers.LinkType) (sniffer.Worker, error) {
	return func(dl layers.LinkType) (sniffer.Worker, error) {
		var icmp4 icmp.ICMPv4Processor
		var icmp6 icmp.ICMPv6Processor
		config, err := cfg.ICMP()
		if err != nil {
			return nil, err
		}
		if config.Enabled() {
			reporter, err := publisher.CreateReporter(config)
			if err != nil {
				return nil, err
			}

			icmp, err := icmp.New(false, reporter, watcher, config)
			if err != nil {
				return nil, err
			}

			icmp4 = icmp
			icmp6 = icmp
		}

		tcp, err := tcp.NewTCP(protocols)
		if err != nil {
			return nil, err
		}

		udp, err := udp.NewUDP(protocols)
		if err != nil {
			return nil, err
		}

		worker, err := decoder.New(flows, dl, icmp4, icmp6, tcp, udp)
		if err != nil {
			return nil, err
		}

		return worker, nil
	}
}
