package main

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
)

var currentProc *process.Process

func buildDrillDownOpts(entity interface{}) (opts []string) {
	switch v := entity.(type) {
	case nil:
		return []string{"shitsfucked"}
	case Processes:
		for _, p := range v {
			if n, _ := p.NumFDs(); n > 0 {
				opts = append(opts, OptName(p))
			}
		}
	case *process.Process:
		if v == nil {
			return []string{"shitsfucked"}
		}
		n, err := v.Username()
		if err == nil {
			u, found := users.ByName(n)
			if found {
				opts = append(opts, OptName(u))
			}
		}

		conns, err := v.Connections()
		for _, c := range conns {
			opts = append(opts, OptName(c))
		}

		ofs, err := v.OpenFiles()
		if err == nil {
			opts = append(opts, buildDrillDownOpts(ofs)...)
		}

	case Users:
		for _, u := range v {
			opts = append(opts, OptName(u))
		}

	case *User:
		ps := procs.ByUser(v.name)
		opts = buildDrillDownOpts(ps)
		opts = append(opts, buildDrillDownOpts(v.OpenFiles)...)

	case []process.OpenFilesStat:
		var fs Files
		for _, of := range v {
			if f, ok := files.ByName(of.Path); ok {
				fs = append(fs, f)
			}
		}
		return buildDrillDownOpts(fs)

	case Files:
		sort.Sort(sort.Reverse(ByFileOpenFiles{v}))
		for _, f := range v {
			opts = append(opts, OptName(f))
		}

	case *File:
		return buildDrillDownOpts(procs.ByFilename(v.path))

	default:
		opts = []string{"Go Back"}
	}
	return
}

func OptName(x interface{}) string {

	switch v := x.(type) {
	case *File:
		open := fmt.Sprintf("(%d times)", v.numFDs)
		return fmt.Sprintf("File: %10s %s", open, v.path)
	case *User:
		return fmt.Sprintf("User: %s (%d open)", v.name, v.numFDs)
	case *process.Process:
		n, _ := v.NumFDs()
		nn, _ := v.Name()
		// e, _ := p.Exe()
		open := fmt.Sprintf("(%d open)", n)

		max := 60
		if len(nn) <= max {
			max = len(nn)
		}
		nn = nn[0:max]

		return fmt.Sprintf("%-5d %10s %10s %s", v.Pid, open, yolo(v.Username()), nn)
	case net.ConnectionStat:
		return fmt.Sprintf("Net: %s:%d", v.Raddr.IP, v.Raddr.Port)
	}

	return "wtf"
}

func ParsePID(x string) int {
	pidS := strings.Split(strings.TrimSpace(x), " ")[0]
	pid, _ := strconv.Atoi(pidS)
	return pid
}

func getDrillToEntity(drillto string) (interface{}, bool) {
	switch {
	case strings.HasPrefix(drillto, "File:"):
		drillto = collapseSpaces(drillto)
		bits := strings.Split(drillto, " ")
		if len(bits) < 4 {
			return nil, false
		}
		fn := bits[3]
		log.Println("Finding file", fn, "x", drillto)
		return files.ByName(fn)

	case strings.HasPrefix(drillto, "User:"):
		n := strings.Split(drillto, " ")[1]
		return users.ByName(n)

	default:
		return procs.ByPID(ParsePID(drillto))
	}
}

func collapseSpaces(s string) string {
	re := regexp.MustCompile(`\s+`)
	return re.ReplaceAllString(s, " ")
}

func drillDownSummary(x interface{}) string {
	switch v := x.(type) {
	case Processes:
		return fmt.Sprintf("Found %d running processes", len(v))
	case *File:
		return fmt.Sprintf("File %s, opened %d times", v.path, v.numFDs)
	case *User:
		return fmt.Sprintf("User %s, %d running processes, %d open files", v.name, len(procs.ByUser(v.name)), len(v.OpenFiles))
	case *process.Process:
		return fmt.Sprintf("Process %d, (%s), run by %s, %d open files", v.Pid, yolo(v.Exe()), yolo(v.Username()), yolo32(v.NumFDs()))
	case Files:
		return fmt.Sprintf("%d files open", len(v))
	case Users:
		return fmt.Sprintf("%d users using files", len(v))
	}
	return "wtf is this"
}

func yolo(s string, _ error) interface{} {
	return s
}

func yolo32(s int32, _ error) interface{} {
	return s
}
