package main

import (
	"code.cloudfoundry.org/cli/plugin"
	"github.com/seattle-beach/cf-cli-rds-plugin/cf_rds"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
)

type MyConfig struct {}

func (c MyConfig) ColorEnabled() configv3.ColorSetting {
	return configv3.ColorAuto
}

func (c MyConfig) Locale() string {
	return "EN-US"
}

func (c MyConfig) IsTTY() bool {
	return false
}

func (c MyConfig) TerminalWidth() int {
	return 192
}

// Unlike most Go programs, the `Main()` function will not be used to run all of the
// commands provided in your plugin. Main will be used to initialize the plugin
// process, as well as any dependencies you might require for your
// plugin.
func main() {
	// Any initialization for your plugin can be handled here
	//
	// Note: to run the plugin.Start method, we pass in a pointer to the struct
	// implementing the interface defined at "code.cloudfoundry.org/cli/plugin/plugin.go"
	//
	// Note: The plugin's main() method is invoked at install time to collect
	// metadata. The plugin will exit 0 and the Run([]string) method will not be
	// invoked.
	config := MyConfig {}
	my_ui, _ := ui.NewUI(&config)
	rds_plugin := cf_rds.BasicPlugin{ UI: my_ui}
	plugin.Start(&rds_plugin)
	// Plugin code should be written in the Run([]string) method,
	// ensuring the plugin environment is bootstrapped.
}
