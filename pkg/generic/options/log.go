package options

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/pflag"
	"k8s.io/component-base/config"
	"k8s.io/component-base/logs"
	"k8s.io/component-base/logs/registry"
	"strings"
)

type LoggingConfiguration struct {
	// Refer [Logs Options](https://github.com/kubernetes/component-base/blob/master/logs/options.go) for more information.
	config.LoggingConfiguration
}

func NewDefaultLoggingConfiguration() LoggingConfiguration {
	return LoggingConfiguration{
		config.LoggingConfiguration{
			Format:    "text",
			Verbosity: 2,
		},
	}
}

func (l *LoggingConfiguration) ValidateAndApply() error {
	o := logs.NewOptions()
	o.Config.Format = l.Format
	o.Config.Verbosity = l.Verbosity
	o.Config.VModule = l.VModule
	return o.ValidateAndApply()
}

type marshalLoggingConfig struct {
	Format    string
	Verbosity config.VerbosityLevel
	VModule   config.VModuleConfiguration
}

func (l *LoggingConfiguration) MarshalJSON() ([]byte, error) {
	return json.Marshal(&marshalLoggingConfig{
		Format:    l.Format,
		Verbosity: l.Verbosity,
		VModule:   l.VModule,
	})
}

func (l *LoggingConfiguration) UnmarshalJSON(bytes []byte) error {
	in := &marshalLoggingConfig{}
	if err := json.Unmarshal(bytes, in); err != nil {
		return err
	}
	l.Format = in.Format
	l.Verbosity = in.Verbosity
	l.VModule = in.VModule
	return nil
}

func (l *LoggingConfiguration) BindLoggingFlags(fs *pflag.FlagSet) {
	// flagSet := flag.NewFlagSet("logging-file", flag.ContinueOnError)
	// klog.InitFlags(flagSet)
	//
	// fs.AddGoFlagSet(flagSet)
	notHidden := map[string]bool{
		"v":              true,
		"vmodule":        true,
		"logging-format": true,
	}

	logsFs := pflag.NewFlagSet("", pflag.ContinueOnError)
	logs.BindLoggingFlags(&l.LoggingConfiguration, logsFs)
	logsFs.VisitAll(func(f *pflag.Flag) {
		if notHidden[f.Name] {
			if f.Name == "logging-format" {
				formats := fmt.Sprintf(`"%s"`, strings.Join(registry.LogRegistry.List(), `", "`))
				f.Usage = fmt.Sprintf("Sets the log format. Permitted formats: %s.", formats)
			}
			return
		}
		f.Hidden = true
	})

	fs.AddFlagSet(logsFs)
	// fs.SetNormalizeFunc(cliflag.WordSepNormalizeFunc)
}
