package logs

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apimachinery/pkg/util/wait"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/config"
	"k8s.io/component-base/logs/registry"
	"k8s.io/klog/v2"
	"math"
	"os"
	"sigs.k8s.io/yaml"
	"strconv"
	"strings"
	"time"
)

type LoggingConfiguration struct {
	Format         string               `json:"format"`
	FlushFrequency time.Duration        `json:"flushFrequency"`
	Verbosity      uint32               `json:"verbosity"`
	File           string               `json:"file"`
	FileMaxSize    int                  `json:"fileMaxSize"`
	Options        config.FormatOptions `json:"options,omitempty"`
}

// BaseOptions has klog format parameters
type BaseOptions struct {
	Config LoggingConfiguration
}

// NewOptions return new klog options
func NewOptions() BaseOptions {
	o := BaseOptions{
		Config: LoggingConfiguration{
			Format:         "text",
			FlushFrequency: 5 * time.Second,
			Verbosity:      2,
			File:           "./gateway.log",
			FileMaxSize:    1800,
		},
	}
	return o
}

func (o *BaseOptions) ValidateAndApply() error {
	errs := o.validate()
	if len(errs) > 0 {
		return utilerrors.NewAggregate(errs)
	}
	o.apply()
	return nil
}

func (o *BaseOptions) validate() []error {
	errs := ValidateLoggingConfiguration(&o.Config, nil)
	if len(errs) != 0 {
		return errs.ToAggregate().Errors()
	}
	return nil
}

func (o *BaseOptions) AddFlags(fs *pflag.FlagSet) {
	BindLoggingFlags(&o.Config, fs)
}

func (o *BaseOptions) apply() {

	// if log format not exists, use nil loggr
	factory, _ := registry.LogRegistry.Get(o.Config.Format)
	if factory == nil {
		klog.ClearLogger()
	} else {
		log, flush := factory.Create(o.Config.Options)
		klog.SetLogger(log)
		logrFlush = flush
	}
	//
	// logFile := "./log/harnsgateway.log"
	// if len(o.Config.File) > 0 {
	// 	logFile = o.Config.File
	// }
	//
	// fileMaxSizeMB := 1800
	// if o.Config.FileMaxSize > 0 {
	// 	fileMaxSizeMB = o.Config.FileMaxSize
	// }
	// klog.SetOutput()

	if err := loggingFlags.Lookup("v").Value.Set(strconv.Itoa(int(o.Config.Verbosity))); err != nil {
		panic(fmt.Errorf("internal error while setting klog verbosity: %v", err))
	}
	go wait.Forever(FlushLogs, o.Config.FlushFrequency)
}
func (o *BaseOptions) AddBaseFlags(cmd *cobra.Command, fs *pflag.FlagSet) {
	o.bindLoggingFlags(fs)
	addHelpAndUsage(cmd, fs)
	addDefaultConfig(fs)
}

func (o *BaseOptions) bindLoggingFlags(fs *pflag.FlagSet) {
	notHidden := map[string]bool{
		"v":              true,
		"logging-format": true,
	}

	logsFs := pflag.NewFlagSet("", pflag.ContinueOnError)
	BindLoggingFlags(&o.Config, logsFs)
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
}

func ValidateLoggingConfiguration(c *LoggingConfiguration, fldPath *field.Path) field.ErrorList {
	errs := field.ErrorList{}
	if c.Format != DefaultLogFormat {
		allFlags := UnsupportedLoggingFlags(cliflag.WordSepNormalizeFunc)
		for _, f := range allFlags {
			if f.DefValue != f.Value.String() {
				errs = append(errs, field.Invalid(fldPath.Child("format"), c.Format, fmt.Sprintf("Non-default format doesn't honor flag: %s", f.Name)))
			}
		}
	}
	_, err := registry.LogRegistry.Get(c.Format)
	if err != nil {
		errs = append(errs, field.Invalid(fldPath.Child("format"), c.Format, "Unsupported log format"))
	}

	// The type in our struct is uint32, but klog only accepts positive int32.
	if c.Verbosity > math.MaxInt32 {
		errs = append(errs, field.Invalid(fldPath.Child("verbosity"), c.Verbosity, fmt.Sprintf("Must be <= %d", math.MaxInt32)))
	}
	return errs
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
