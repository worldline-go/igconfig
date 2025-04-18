package igconfig_test

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/worldline-go/igconfig"
	"github.com/worldline-go/igconfig/loader"
	"github.com/worldline-go/igconfig/testdata"
)

// Example_fileLoader for getting values from file(YAML or JSON).
// In this example, used to change etc path and get the file name from the appname.
func Example_fileLoader() {

	/* ===== YAML file of testdata/config/train.yml =====
	      speed: 10
	      ADDRESS: "Hoofddrop"
	      bay: 4
	      secure: false
	      info:
	        train_name: "IN-NS1234"
	        age: 10
	        destination: "Eindhoven"
	      InfoStruct:
	        train_name: "Embedded-NS1234"
	        # WARNING RANDOM GETS IF SAME KEY EXISTS
	   	 # testing | TesTing | tesTing are same for Testing | TesTing fields in struct
	        testing: "testX"
	        # TesTing: "testY"
	        # tesTing: "testZ"
	        # DoubleName: "textD"
	        # doubleName: "textD"
	        doublename: "textD"
	      extra:
	        new: "Extra data ignore"
	*/

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

	mytrain := AwesomeTrain{}

	// If not found, it return an error in igconfig loader.
	// os.Setenv(loader.EnvConfigFile, "testdata/config/train.yml")
	// defer os.Unsetenv(loader.EnvConfigFile)

	// err := igconfig.LoadConfig("train", &mytrain)
	// For this example just used file loader, most cases just use LoadConfig function.
	err := igconfig.LoadWithLoaders("train", &mytrain, loader.File{EtcPath: "testdata/config"})
	if err != nil {
		log.Fatal("unable to load configuration settings", err)
	}

	fmt.Printf("%+v", mytrain)

	// Output:
	// {Name: Age:0 Speed:10 Address:Hoofddrop Bay:4 Secure:false Slice:[] InfoStruct:{Name:Embedded-NS1234 Age:0 Destination: Testing:testX TesTing:testX DoubleName:textD Time:0001-01-01 00:00:00 +0000 UTC} Info:{Name:IN-NS1234 Age:10 Destination:Eindhoven Testing: TesTing: DoubleName: Time:0001-01-01 00:00:00 +0000 UTC} InfoStructSkip:{Name: Age:0 Destination: Testing: TesTing: DoubleName: Time:0001-01-01 00:00:00 +0000 UTC}}
}

func ExampleLoadWithLoaders() {
	// If only particular loaders are needed or new loader should be added - it is possible to do.
	//
	// igconfig.DefaultLoaders is an array of loaders provided by default.
	//
	// This example uses only Flags loader.
	// This means that no default or environmental variables will be loaded.
	//
	// Some loaders may accept additional configuration when used like this
	flagsLoader := loader.Flags{
		NoUsage: true,
	}

	// Prepare pre-defined list of flags for this example
	os.Args = []string{"executable", "-salary", "12345.66"}

	var c testdata.TestConfig

	// igconfig.LoadWithLoaders provides ability to use specific loaders.
	//
	// P.S.: Please check errors in your code.
	_ = igconfig.LoadWithLoaders("adm0001s", &c, flagsLoader)

	fmt.Println(c.Name)
	fmt.Println(c.Salary)

	// Output:
	//
	// 12345.66
}
