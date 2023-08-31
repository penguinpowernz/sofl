package main

import (
	"sort"

	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
)

type Processes []*process.Process

func (ps Processes) Len() int      { return len(ps) }
func (ps Processes) Swap(i, j int) { ps[i], ps[j] = ps[j], ps[i] }

func (procs Processes) ByPID(pid int) (*process.Process, bool) {
	for _, p := range procs {
		if p.Pid == int32(pid) {
			return p, true
		}
	}
	return nil, false
}

func (procs Processes) ByUser(name string) Processes {
	var res Processes
	for _, p := range procs {
		u, _ := p.Username()
		if u == name {
			res = append(res, p)
		}
	}
	return res
}

func (procs Processes) ByFilename(name string) Processes {
	var res Processes
	for _, p := range procs {
		ofs, err := p.OpenFiles()
		if err != nil {
			continue
		}

		for _, s := range ofs {
			if s.Path == name {
				res = append(res, p)
				break
			}
		}
	}
	return res
}

type ByProcOpenFiles struct{ Processes }

func (s ByProcOpenFiles) Less(i, j int) bool {
	ni, _ := s.Processes[i].NumFDs()
	nj, _ := s.Processes[j].NumFDs()
	return ni < nj
}

func howManyTimesDoesThisProcessHaveThisFileOpened(process *process.Process, filename string) int {
	ofs, err := process.OpenFiles()
	if err != nil {
		return 0
	}

	list := perFileOpenCounts(ofs)
	i := list.Index(filename)
	if i == -1 {
		return 0
	}

	return list[i].Value
}

func perFileOpenCounts(ofs []process.OpenFilesStat) PairList {
	var list PairList

	for _, of := range ofs {
		i := list.Index(of.Path)
		if i == -1 {
			list = append(list, Pair{of.Path, 1})
			continue
		}

		list[i].Value++
	}

	sort.Sort(sort.Reverse(list))
	return list
}

func perSocketOpenCounts(cs []net.ConnectionStat) PairList {
	var list PairList

	for _, c := range cs {
		i := list.Index(connAddr(c))
		if i == -1 {
			list = append(list, Pair{connAddr(c), 1})
			continue
		}

		list[i].Value++
	}

	sort.Sort(sort.Reverse(list))
	return list
}

type Pair struct {
	Key   string
	Value int
}

type PairList []Pair

func (list PairList) Index(key string) int {
	for i, pair := range list {
		if pair.Key == key {
			return i
		}
	}
	return -1
}

func (p PairList) Len() int           { return len(p) }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
