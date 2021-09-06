package igconfig_test

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2/loader"

	"github.com/rs/zerolog"

	"gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2"

	"github.com/stretchr/testify/assert"

	"gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2/testdata"
)

func ExampleLoadConfig() {
	var config testdata.TestConfig

	// Disable logging for unreachable local services.
	// In non-local environments this should not be done.
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	// Below are just an examples of how values can be provided. You don't need to do this in your code.
	// In real-world - this will be provided from env, flags or Consul/Vault
	os.Args = []string{"executable", "-name", "FromFlags"}
	_ = os.Setenv("PORT", "5647")

	if err := igconfig.LoadConfig("adm0001s", &config); err != nil {
		log.Fatalf("load configuration: %s", err.Error())
	}

	fmt.Println(config.Host) // This value is set from default
	fmt.Println(config.Name) // This value is set from application flags
	fmt.Println(config.Port) // This value is set from environmental variable

	// Output:
	// localhost
	// FromFlags
	// 5647

}

var (
	// AppName includes name of the binary
	AppName string = "test"
	// AppVersion contains version of it(commit id)
	AppVersion string
)

type Inner struct {
	DeepestSecret      string `env:"DEEPEST_SECRET"           secret:"deepest_secret"`
	TopSecret          string `env:"TOP_SECRET"               secret:"top_secret"`
}

// Config struct detailing all project parameters.
type Config struct {
	TestKey string `env:"test_key"           secret:"test_key"`
	Inner   Inner
}

type inner struct {
	Field2 string `secret:"field_2"`
}

type testStruct struct {
	Field1   string `secret:"field_1"`
	Untagged int64
	NoSet    string `secret:"-"`
	NoData   string `secret:"missing"`
	Time     time.Time
	Inner    inner `secret:"other"`
}


func TestLoadConfig2(t *testing.T) {

	_ = os.Setenv("VAULT_ROLE_ID", "dba2ddf3-8c94-8810-1325-1a1c7ad66d88")
	_ = os.Setenv("VAULT_ADDR", "https://am2vm2356.test.igdcs.com:8200")
	os.Args = []string{"program"}
	var config testStruct
	if err := igconfig.LoadConfig(AppName, &config); err != nil {
		log.Fatalf("load configuration: %s", err.Error())
	}
	os.Unsetenv("VAULT_ADDR")
	os.Unsetenv("VAULT_ROLE_ID")


	fmt.Println("Actual config", config)
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

func TestLoadConfig(t *testing.T) {
	c1 := testdata.BadDefaults{}

	assert.NotNil(t, igconfig.LoadConfig("haha", &c1))

	var c2 testdata.TestConfig

	prevArgs := os.Args
	os.Args = []string{"program", "--age", "haha", "--salary", "nothing"}

	assert.NotNil(t, igconfig.LoadConfig("", &c2))
	os.Args = prevArgs
}
