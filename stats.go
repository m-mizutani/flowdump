package main

import (
	"time"

	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
)

type connection struct {
	localAddr  string
	remoteAddr string
	localPort  uint32
	remotePort uint32
	pid        uint
	state      string
	pname      string
}

type connList []connection

func startNetStat(interval int) (chan connList, chan error) {
	connCh := make(chan connList)
	errCh := make(chan error)

	go func() {
		defer close(connCh)
		defer close(errCh)

		for {
			conns, err := net.Connections("")

			if err != nil {
				errCh <- err
				return
			}

			procCache := map[int32]*process.Process{}

			connStats := connList{}
			for _, conn := range conns {
				if conn.Status == "CLOSED" || conn.Status == "LISTEN" || conn.Status == "" {
					continue
				}

				connStat := connection{
					localAddr:  conn.Laddr.IP,
					localPort:  conn.Laddr.Port,
					remoteAddr: conn.Raddr.IP,
					remotePort: conn.Raddr.Port,
					state:      conn.Status,
					pid:        uint(conn.Pid),
				}

				proc, ok := procCache[conn.Pid]
				if !ok {
					proc, err = process.NewProcess(conn.Pid)
					if err == nil && proc != nil {
						procCache[conn.Pid] = proc
					}
				}

				pname, err := proc.Name()
				if err != nil {
					errCh <- err
					return
				}
				connStat.pname = pname
				connStats = append(connStats, connStat)
			}
			connCh <- connStats

			time.Sleep(time.Millisecond * time.Duration(interval))
		}
	}()

	return connCh, errCh
}
