package igconfig

import (
	"gitlab.test.igdcs.com/finops/utils/basics/iglog.git"
	"reflect"
	"testing"
)

type ValueConfig struct {
	LogLevel      string `env:"loglevel" cmd:"loglevel,log,g"    cfg:"loglevel,log" default:"info"`
	ReportID      int    `env:"reportid" cmd:"reportid,report,r"                    default:"100"`
	LBAnr         int    `env:"lbanr"    cmd:"lbanr,lba,l"                          default:"0"`
}

func TestIsTrue(t *testing.T) {
	testcases := []struct {
		c string
		w bool
	}{
		{"TRUE",  true},
		{"true",  true},
		{"T",     true},
		{"t",     true},
		{".t.",   true},
		{"YeS",   true},
		{"1",     true},
		{"FALSE", false},
		{"treu",  false},
		{"tru",   false},
		{"0",     false},
	}

	for _, testcase := range testcases {
		g := isTrue(testcase.c)
		if g != testcase.w {
			t.Errorf("TestIsTrue failed; got=%t; want=%t", g, testcase.w)
		}
	}
}

func TestSetValue(t *testing.T) {
	var c ValueConfig

	fLogLevel, ok := reflect.TypeOf(&c).Elem().FieldByName("LogLevel")
	if !ok {
		t.Error("TestSetValue cannot get reflect field for LogLevel")
	}

	s := "error"
	l := iglog.LogError

	setValue(&c, fLogLevel, s)
	if c.LogLevel != s {
		t.Errorf("TestSetValue failed to set LogLevel; got=%s; want=%s", c.LogLevel, s)
	}

	if iglog.Level() != l {
		t.Errorf("TestSetValue global LogLevel not set properly; got=%s; want=%s", iglog.Level().String(), l.String())
	}
}