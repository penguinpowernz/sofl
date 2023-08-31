package main

import (
	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
)

type User struct {
	name      string
	numFDs    int
	OpenFiles []process.OpenFilesStat
	OpenConns []net.ConnectionStat
}

type Users []*User

func (us Users) Len() int      { return len(us) }
func (us Users) Swap(i, j int) { us[i], us[j] = us[j], us[i] }

func (us Users) ByName(n string) (*User, bool) {
	for _, u := range us {
		if u.name == n {
			return u, true
		}
	}

	return nil, false
}

func (u *User) AddUniqueOpenFiles(ofs []process.OpenFilesStat) {
	for _, of := range ofs {
		var found bool
		for _, of1 := range u.OpenFiles {
			if of1.Path == of.Path {
				found = true
				break
			}
		}
		if !found {
			u.OpenFiles = append(u.OpenFiles, of)
		}
	}
	u.numFDs = len(u.OpenFiles)
}

type ByUserOpenFiles struct{ Users }

func (s ByUserOpenFiles) Less(i, j int) bool { return s.Users[i].numFDs < s.Users[j].numFDs }
