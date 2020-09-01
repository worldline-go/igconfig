package testdata

type TestConfig struct {
	Name    string  `cfg:"settle_name"    env:"name"           cmd:"name,n"           default:"Jan"`
	Age     uint    `cfg:"age"            env:"age"            cmd:"age,a"            default:"18"`
	Salary  float64 `cfg:"salary"         env:"salary"         cmd:"salary,s"         default:"2000.00"  loggable:"false"`
	Host    string  `cfg:"host"           default:"localhost"`
	Address string  `cfg:"ADDRESS"        env:"ADDRESS"        default:"localhost"`
	Port    int     `cfg:"port"           cmd:"port,p"           default:"8080"`
	Secure  bool    `cfg:"secure,ssl,tls" env:"secure,ssl,tls" cmd:"secure,ssl,tls,t" default:"false"    loggable:"false"`
	Unused  []string
}

type BadDefaults struct {
	Age    uint    `cfg:"age"            env:"age"            cmd:"age,a"            default:"haha"`
	Salary float64 `cfg:"salary"         env:"salary"         cmd:"salary,s"         default:"haha"`
	Port   int     `cfg:"port"           env:"port"           cmd:"port,p"           default:"haha"`
}
