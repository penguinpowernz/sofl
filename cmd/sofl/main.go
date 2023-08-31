package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sort"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/shirou/gopsutil/process"
)

var (
	users Users
	files Files
	procs Processes
)

func main() {

	var pid int
	flag.IntVar(&pid, "p", 0, "process to display report for")
	flag.Parse()

	if pid > 0 {
		getProcessState()
		p, found := procs.ByPID(pid)
		if !found {
			fmt.Printf("Couldn't find process with PID %d\n", pid)
			os.Exit(1)
		}

		io.Copy(os.Stdout, buildProcessReport(p))
		os.Exit(0)
	}

	getProcessState()
	go func() {
		for {
			getProcessState()
			time.Sleep(time.Second * 2)
		}
	}()

	for {

		var actn string
		err := survey.AskOne(&survey.Select{
			Message: "What to do:",
			Options: []string{"Explore", "Processes"},
		}, &actn)
		if err != nil {
			log.Println(err)
			break
		}

		switch actn {
		case "Processes":
			for {
				opts := buildDrillDownOpts(procs)

				var proc string
				err = survey.AskOne(&survey.Select{
					Message:  "Inspect process:",
					Options:  opts,
					PageSize: 30,
				}, &proc)
				if err != nil {
					log.Println(err)
					break
				}

				pid := ParsePID(proc)
				p, found := procs.ByPID(pid)
				if !found {
					fmt.Printf("Couldn't find process with PID %d\n", pid)
					continue
				}

				buf := buildProcessReport(p)
				less(buf)
			}

		case "Explore":
			for {
				var startWith string
				err := survey.AskOne(&survey.Select{
					Message: "Start with:",
					Options: []string{"Processes", "Files", "Users"},
				}, &startWith)
				if err != nil {
					log.Println(err)
					break
				}

				var prevEntity interface{}
				var entity interface{}

				switch startWith {
				case "Processes":
					entity = procs
				case "Files":
					entity = files
				case "Users":
					entity = users

				}

				prevEntity = entity

				for {
					opts := buildDrillDownOpts(entity)

					fmt.Print("\n", drillDownSummary(entity), "\n")

					var drillto string
					err = survey.AskOne(&survey.Select{
						Message:  "Drill Down:",
						Options:  opts,
						PageSize: 30,
					}, &drillto)
					if err != nil {
						log.Println(err)
						break
					}

					if drillto == "Go Back" {
						entity = prevEntity
						continue
					}

					var found bool
					prevEntity = entity
					entity, found = getDrillToEntity(drillto)
					if !found {
						fmt.Println("didnt find that guy")
						entity = prevEntity
						continue
					}
				}
			}
		}
	}

	// sort.Sort(ByUserOpenFiles{users})
	// for _, u := range users {
	// 	fmt.Println(u.numFDs, u.name)
	// }

	// sort.Sort(ByFileOpenFiles{files})
	// for _, f := range files {
	// 	fmt.Println(f.numFDs, f.path)
	// }

}

func buildProcessReport(p *process.Process) *bytes.Buffer {
	buf := bytes.NewBufferString(fmt.Sprintf("Process: %d\n", p.Pid))
	buf.WriteString(fmt.Sprintf("Exe: %s\n", yolo(p.Exe())))
	buf.WriteString(fmt.Sprintf("Cmdline: %s\n", yolo(p.Cmdline())))
	buf.WriteString("\n")
	buf.WriteString("Network connections:")
	buf.WriteString("\n")

	conns, _ := p.Connections()
	counts := perSocketOpenCounts(conns)
	for _, p := range counts {
		buf.WriteString(fmt.Sprintf("%5d %s\n", p.Value, p.Key))
	}

	buf.WriteString("\n")
	ofs, _ := p.OpenFiles()
	counts = perFileOpenCounts(ofs)
	buf.WriteString("Open file counts:")
	buf.WriteString("\n")
	for _, p := range counts {
		buf.WriteString(fmt.Sprintf("%5d %s\n", p.Value, p.Key))
	}

	return buf
}

func getProcessState() (Users, Processes, Files) {
	var err error

	procs, err = process.Processes()
	if err != nil {
		panic(err)
	}

	for _, p := range procs {
		ofs, err := p.OpenFiles()
		if err != nil {
			continue
		}

		n, err := p.Username()
		if err == nil {
			u, found := users.ByName(n)
			if !found {
				u = new(User)
				u.name = n
				users = append(users, u)
			}
			u.AddUniqueOpenFiles(ofs)
		}

		files.RegisterOpenFiles(ofs)
	}

	sort.Sort(sort.Reverse(ByProcOpenFiles{procs}))
	sort.Sort(sort.Reverse(ByUserOpenFiles{users}))

	return users, procs, files
}

func less(r io.Reader) {
	cmd := exec.Command("less")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = r
	cmd.Run()
}
