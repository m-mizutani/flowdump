package main

import (
	"fmt"

	"github.com/k0kubun/pp"
	"github.com/shirou/gopsutil/net"
)

func main() {
	conn, err := net.Connections("")
	pp.Println(conn)
	if err != nil {
		fmt.Println("error", err)
	}
}
