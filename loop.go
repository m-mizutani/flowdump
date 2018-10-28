package main

import (
	"encoding/binary"
	"hash/fnv"
	"net"

	"github.com/k0kubun/pp"
)

type flowDumpOpts struct {
	deviceName      string
	intervalNetStat int
	intervalCapture int
}

func newFlowDumpOpts() flowDumpOpts {
	opts := flowDumpOpts{}
	opts.intervalCapture = 1000
	opts.intervalNetStat = 1000
	return opts
}

func calcHash(addr1, port1, addr2, port2 []byte) uint32 {
	d := make([]byte, 0, len(addr1)+len(addr2)+len(port1)+len(port2))
	d = append(d, addr1...)
	d = append(d, port1...)
	d = append(d, addr2...)
	d = append(d, port2...)
	hash := fnv.New32()
	hash.Sum(d)
	return hash.Sum32()
}

func hashFlow(a1, a2 net.IP, p1, p2 uint32) uint32 {
	pb1 := make([]byte, 4)
	pb2 := make([]byte, 4)
	binary.LittleEndian.PutUint32(pb1, p1)
	binary.LittleEndian.PutUint32(pb2, p2)

	hv1 := calcHash(a1, pb1, a2, pb2)
	hv2 := calcHash(a2, pb2, a1, pb1)
	if hv1 < hv2 {
		return hv1
	}

	return hv2
}

func hashConn(conn connection) uint32 {
	la := net.ParseIP(conn.localAddr)
	ra := net.ParseIP(conn.remoteAddr)

	return hashFlow(la, ra, conn.localPort, conn.remotePort)
}

type connCache map[uint32]connection

func loop(opts flowDumpOpts) error {
	netCh, netErr := startNetStat(opts.intervalNetStat)
	flowCh, flowErr := startCapture(opts.deviceName, opts.intervalCapture)
	cache := connCache{}

	for {
		select {
		case connections := <-netCh:
			for _, conn := range connections {
				hv := hashConn(conn)
				cache[hv] = conn
			}

		case err := <-netErr:
			return err

		case flows := <-flowCh:
			pp.Println(flows)

		case err := <-flowErr:
			return err
		}
	}
}
