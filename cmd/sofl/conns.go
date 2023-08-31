package main

import (
	"fmt"

	"github.com/shirou/gopsutil/net"
)

type Conn struct {
	addr   string
	numFDs int
}

type Conns []*Conn

func (fs Conns) Len() int      { return len(fs) }
func (fs Conns) Swap(i, j int) { fs[i], fs[j] = fs[j], fs[i] }

func (fs Conns) ByName(n string) (*Conn, bool) {
	for _, f := range fs {
		if f.addr == n {
			return f, true
		}
	}

	return nil, false
}

func (fs *Conns) RegisterOpenFiles(cns []net.ConnectionStat) {
	for _, c := range cns {
		var found bool
		for _, f := range *fs {
			if f.addr == connAddr(c) {
				found = true
				f.numFDs++
			}
		}
		if !found {
			*fs = append(*fs, &Conn{addr: connAddr(c), numFDs: 1})
		}
	}
}

type ByConnOpenFiles struct{ Conns }

func (s ByConnOpenFiles) Less(i, j int) bool {
	return s.Conns[i].numFDs < s.Conns[j].numFDs
}

func connAddr(c net.ConnectionStat) string {
	return fmt.Sprintf("%s:%d", c.Raddr.IP, c.Raddr.Port)
}
