package testdata

import "time"

const Time = "2000-01-01T10:00:00Z"

var ParsedTime = MustParseTime(nil, Time)

type InnerStruct struct {
	Str  string        `cfg:"string" default:"val"`
	Dur  time.Duration `cmd:"dur"`
	Time time.Time     `cfg:"time" default:"2000-01-01T10:00:00Z"`
}

type TestConfig struct {
	Name            string        `cfg:"settle_name"    env:"name"           cmd:"name,n"           default:"Jan"`
	Age             uint          `cfg:"age"            env:"age"            cmd:"age,a"            default:"18"`
	Salary          float64       `cfg:"salary"         env:"salary"         cmd:"salary,s"         default:"2000.00"  loggable:"false"`
	Host            string        `cfg:"host"           default:"localhost"`
	Address         string        `cfg:"ADDRESS"        env:"ADDRESS"        default:"localhost"`
	Port            int           `cfg:"port"           cmd:"port,p"           default:"8080"`
	Secure          bool          `cfg:"secure" env:"secure" cmd:"secure" default:"false"    loggable:"false"`
	Slice           []string      `cfg:"slice" env:"slice" cmd:"slice" default:"1,2,5,6"`
	Dur             time.Duration `cfg:"dur"`
	InnerStruct     InnerStruct
	InnerStructSkip InnerStruct `cfg:"-" default:"-"`
}

type BadDefaults struct {
	Age    uint    `cfg:"age"            env:"age"            cmd:"age,a"            default:"haha"`
	Salary float64 `cfg:"salary"         env:"salary"         cmd:"salary,s"         default:"haha"`
	Port   int     `cfg:"port"           env:"port"           cmd:"port,p"           default:"haha"`
}

type UntaggedInnerStruct struct {
	Str  string
	Dur  time.Duration
	Time time.Time
}

type UntaggedTestConfig struct {
	Name            string
	Age             uint
	Salary          float64
	Host            string
	InnerStruct     UntaggedInnerStruct
	InnerStructSkip InnerStruct `cfg:"-" default:"-"`
}
