package options

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/klog/v2"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
)

type Optioner interface {
	AddFlags(*pflag.FlagSet)
	GetBaseOptions() *BaseOptions
}

type BaseOptions struct {
	ConfigFile string `json:"-"`
	Logging    LoggingConfiguration
}

func NewDefaultBaseOptions() BaseOptions {
	return BaseOptions{
		Logging: NewDefaultLoggingConfiguration(),
	}
}

func (bo *BaseOptions) GetBaseOptions() *BaseOptions {
	return bo
}

func (bo *BaseOptions) AddBaseFlags(cmd *cobra.Command, fs *pflag.FlagSet) {
	bo.addConfigFile(fs)
	bo.addLogging(fs)
	addHelpAndUsage(cmd, fs)
	addDefaultConfig(fs)
}

func (bo *BaseOptions) addConfigFile(fs *pflag.FlagSet) {
	fs.StringVarP(&bo.ConfigFile, "config", "c", bo.ConfigFile, "The program will load its initial configuration from this file. The path may be absolute or relative; relative paths start at the program current working directory. Omit this flag to use the built-in default configuration values. Command-line flags override configuration from this file.")
}

func (bo *BaseOptions) addLogging(fs *pflag.FlagSet) {
	bo.Logging.BindLoggingFlags(fs)
}

func (bo *BaseOptions) ValidateAndApply() error {
	return bo.Logging.ValidateAndApply()
}

func PrintHelpAndExitIfRequested(cmd *cobra.Command, fs *pflag.FlagSet) {
	help, err := fs.GetBool("help")
	if err != nil {
		klog.InfoS(`"help" flag is non-bool, programmer error, please correct`)
		os.Exit(1)
	}
	if help {
		_ = cmd.Help()
		os.Exit(0)
	}
}

func addDefaultConfig(fs *pflag.FlagSet) {
	fs.Bool("default-config", false, "print default configuration for reference, users can refer to it to create their own configuration files")
}

func PrintDefaultConfigAndExitIfRequested(config interface{}, fs *pflag.FlagSet) {
	defaultConfig, err := fs.GetBool("default-config")
	if err != nil {
		klog.InfoS(`"defaultConfig" flag is non-bool, programmer error, please correct`)
		os.Exit(1)
	}
	if defaultConfig {
		data, err := yaml.Marshal(config)
		if err != nil {
			klog.ErrorS(err, "Failed to marshal default config to yaml")
			os.Exit(1)
		}
		fmt.Println("# With --default-config flag, users can easily get a default full config file as reference, with all fields (and field descriptions) included and default values set. ")
		fmt.Println("# Users can modify/create their own configs accordingly as reference. ")
		fmt.Printf("\n%v\n\n", string(data))
		os.Exit(0)
	}
}

func addHelpAndUsage(cmd *cobra.Command, fs *pflag.FlagSet) {
	fs.BoolP("help", "h", false, fmt.Sprintf("help for %s", cmd.Name()))

	// ugly, but necessary, because Cobra's default UsageFunc and HelpFunc pollute the flagset with global flags
	const usageFmt = "Usage:\n  %s\n\nFlags:\n%s"
	cmd.SetUsageFunc(func(cmd *cobra.Command) error {
		_, _ = fmt.Fprintf(cmd.OutOrStderr(), usageFmt, cmd.UseLine(), fs.FlagUsagesWrapped(2))
		return nil
	})

	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n"+usageFmt, cmd.Long, cmd.UseLine(), fs.FlagUsagesWrapped(2))
	})
}

func flagPrecedence(o Optioner, args []string) error {
	fs := pflag.NewFlagSet("", pflag.ExitOnError)
	o.AddFlags(fs)
	o.GetBaseOptions().addConfigFile(fs)
	o.GetBaseOptions().addLogging(fs)

	// re-parse flags
	if err := fs.Parse(args); err != nil {
		return err
	}
	return nil
}

func ParseAndApplyConfigFile(o Optioner, args []string) error {
	if len(o.GetBaseOptions().ConfigFile) == 0 {
		return nil
	}

	if err := parseConfigFile(o); err != nil {
		return err
	}
	if err := flagPrecedence(o, args); err != nil {
		return err
	}

	return nil
}

func parseConfigFile(out Optioner) error {
	configFilePath, err := filepath.Abs(out.GetBaseOptions().ConfigFile)
	if err != nil {
		klog.ErrorS(err, "Failed to load config file", "file", configFilePath)
		return err
	}

	data, err := os.ReadFile(configFilePath)
	if err != nil {
		klog.ErrorS(err, "Failed to read config file", "file", configFilePath)
		return err
	}

	err = yaml.Unmarshal(data, out)
	if err != nil {
		klog.ErrorS(err, "Failed to unmarshal config file", "file", configFilePath)
		return err
	}
	return nil
}
