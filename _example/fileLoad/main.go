package main

import (
	"fmt"
	"log"
	"time"

	"gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2"
	// "gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2/loader"
)

type InfoStruct struct {
	Name        string `cfg:"train_name"`
	Age         uint   `cfg:"age"`
	Destination string
	Testing     string
	TesTing     string
	DoubleName  string
	Time        time.Time `cfg:"time" default:"2000-01-01T10:00:00Z"`
}

type AwesomeTrain struct {
	Name    string   `cfg:"train_name"    env:"name"           cmd:"name,n"           default:"NS1"`
	Age     uint     `cfg:"age"            env:"age"            cmd:"age,a"            default:"18"`
	Speed   float64  `cfg:"speed"         env:"speed"         cmd:"speed,s"         default:"200.50"  loggable:"false"`
	Address string   `cfg:"ADDRESS"        env:"ADDRESS"        default:"localhost"`
	Bay     int      `cmd:"bay,p"           default:"10"`
	Secure  bool     `cfg:"secure" env:"secure" cmd:"secure" default:"false"    loggable:"false"`
	Slice   []string `cfg:"slice" env:"slice" cmd:"slice" default:"1,2,5,6"`
	InfoStruct
	Info           InfoStruct
	InfoStructSkip InfoStruct `cfg:"-" default:"-"`
}

// FileLoader get file path on CONFIG_FILE environment variable also it check working directory and /etc folder
// in our example we put a yml file in working directory, file suffix could be yml, yaml or json.

func main() {
	mytrain := AwesomeTrain{}

	// If not found, it return an error in igconfig loader.
	// os.Setenv("CONFIG_FILE", "dsad.yaml")

	// err := igconfig.LoadWithLoaders("train", &mytrain, loader.File{})
	err := igconfig.LoadConfig("train", &mytrain)
	if err != nil {
		log.Fatal("unable to load configuration settings", err)
	}

	fmt.Printf("%+v\n", mytrain)
	fmt.Printf("%v\n", mytrain.Destination)
}
