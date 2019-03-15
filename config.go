package igconfig

import (
	"gitlab.test.igdcs.com/finops/utils/basics/iglog.git"
)

// LoadConfig loads configuration parameters from a file, the environment and finally from
// command-line parameters (the latter override the former) into a config struct.
func LoadConfig(c interface{}, file string, env bool, cmd bool, log bool) error {
	iglog.DebugFull("LoadConfig: start")
	defer iglog.DebugFull("LoadConfig: end")

	err := loadDefaults(c)
	if err != nil {
		return err
	}

	if file != "" {
		err = loadConfigFile(c, file)
	}

	if env {
		loadEnv(c)
	}

	if cmd {
		loadCmdline(c)
	}

	if log && iglog.Level() <= iglog.LogInfo {
		displayConfig(c)
	}

	return err
}
