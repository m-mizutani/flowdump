package main

import (
	"fmt"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type flow interface {
	label() string
	tuples() (net.IP, net.IP, uint32, uint32, string)
}

type baseFlow struct {
	srcAddr net.IP
	dstAddr net.IP
}

type tcpFlow struct {
	baseFlow
	srcPort uint32
	dstPort uint32
}

func (x *tcpFlow) label() string {
	return fmt.Sprintf("%s:%d -> %s:%d", x.srcAddr.String(), x.srcPort, x.dstAddr.String(), x.dstPort)
}

func (x *tcpFlow) tuples() (net.IP, net.IP, uint32, uint32, string) {
	return x.srcAddr, x.dstAddr, x.srcPort, x.dstPort, "TCP"
}

type flowMap map[uint32]flow

func decodePacket(pkt gopacket.Packet) flow {
	ipLayer := pkt.Layer(layers.LayerTypeIPv4)

	if ipLayer != nil {
		ip, _ := ipLayer.(*layers.IPv4)

		tcpLayer := pkt.Layer(layers.LayerTypeTCP)
		if tcpLayer != nil {
			tcp, _ := tcpLayer.(*layers.TCP)

			tcpf := tcpFlow{
				srcPort: uint32(tcp.SrcPort),
				dstPort: uint32(tcp.DstPort),
			}
			tcpf.srcAddr = []byte(ip.SrcIP)
			tcpf.dstAddr = []byte(ip.DstIP)

			return &tcpf
		}
	}

	return nil
}

func startCapture(deviceName string, interval int) (chan flowMap, chan error) {
	flowCh := make(chan flowMap)
	errCh := make(chan error)

	go func() {
		defer close(flowCh)
		defer close(errCh)

		var snapshotLen int32 = 0xffff
		promiscuous := true
		timeout := -1 * time.Second

		handler, err := pcap.OpenLive(deviceName, snapshotLen, promiscuous, timeout)

		if err != nil {
			errCh <- err
			return
		}

		packetSource := gopacket.NewPacketSource(handler, handler.LinkType())
		pktCh := packetSource.Packets()
		delta := time.Duration(interval) * time.Millisecond
		timeoutCh := time.After(delta)
		fmap := flowMap{}

		for {
			select {
			case pkt := <-pktCh:
				f := decodePacket(pkt)
				if f != nil {
					sa, da, sp, dp, _ := f.tuples()
					hv := hashFlow(sa, da, sp, dp)
					fmap[hv] = f
				}

			case <-timeoutCh:
				flowCh <- fmap
				fmap = flowMap{}
				timeoutCh = time.After(delta)
			}
		}
	}()

	return flowCh, errCh
}
