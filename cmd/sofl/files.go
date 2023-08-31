package main

import "github.com/shirou/gopsutil/process"

type File struct {
	path   string
	numFDs int
}

type Files []*File

func (fs Files) Len() int      { return len(fs) }
func (fs Files) Swap(i, j int) { fs[i], fs[j] = fs[j], fs[i] }

func (fs Files) ByName(n string) (*File, bool) {
	for _, f := range fs {
		if f.path == n {
			return f, true
		}
	}

	return nil, false
}

func (fs *Files) RegisterOpenFiles(ofs []process.OpenFilesStat) {
	for _, of := range ofs {
		var found bool
		for _, f := range *fs {
			if f.path == of.Path {
				found = true
				f.numFDs++
			}
		}
		if !found {
			*fs = append(*fs, &File{path: of.Path, numFDs: 1})
		}
	}
}

type ByFileOpenFiles struct{ Files }

func (s ByFileOpenFiles) Less(i, j int) bool {
	return s.Files[i].numFDs < s.Files[j].numFDs
}
