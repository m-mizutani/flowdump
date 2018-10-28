package main

import "github.com/k0kubun/pp"

func dump(flows flowMap, cache connCache) error {
	pp.Println(flows)
	return nil
}
