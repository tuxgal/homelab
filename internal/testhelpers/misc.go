package testhelpers

import (
	"os"

	"github.com/tuxgal/tuxlog"
)

func HomelabBaseDir() string {
	return "testdata/dummy-base-dir"
}

func Pwd() string {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return pwd
}

func NewLogLevel(lvl tuxlog.Level) *tuxlog.Level {
	return &lvl
}

func NewInt(i int) *int {
	return &i
}
