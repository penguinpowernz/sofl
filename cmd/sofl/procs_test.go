package main

import (
	"testing"

	"github.com/shirou/gopsutil/process"
	. "github.com/smartystreets/goconvey/convey"
)

func TestProcessPerFileOpenCounts(t *testing.T) {

	Convey("given an array of open files", t, func() {
		ofs := []process.OpenFilesStat{{Path: "/x"}, {Path: "/x"}, {Path: "/y"}}

		Convey("when counting the open files", func() {
			list := perFileOpenCounts(ofs)

			Convey("it return enough array items", func() {
				So(list, ShouldHaveLength, 2)
			})

			Convey("it should have 2 of /x", func() {
				i := list.Index("/x")
				So(i, ShouldNotEqual, -1)
				So(list[i].Value, ShouldEqual, 2)
			})

			Convey("it should have 1 of /y", func() {
				i := list.Index("/y")
				So(i, ShouldNotEqual, -1)
				So(list[i].Value, ShouldEqual, 1)
			})
		})
	})
}
