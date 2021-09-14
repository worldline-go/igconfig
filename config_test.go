package igconfig_test

import (
	"fmt"
	"log"
	"os"
	"testing"

	"gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2/loader"

	"github.com/rs/zerolog"

	"gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

func TestLoadConfig_Skip(t *testing.T) {
	testLoaders := []loader.Loader{
		loader.File{},
	}

	var c testdata.TestConfig

	require.NoError(t, igconfig.LoadWithLoaders("skipApp", &c, testLoaders...))
}
