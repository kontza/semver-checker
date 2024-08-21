/*
Copyright © 2024 Juha Ruotsalainen <juha.ruotsalainen@iki.fi>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "semver-checker",
	Short:   "Search Gitlab for a package based on name and version",
	Run:     rootRunner,
	Version: "v1.0.0",
	Args:    cobra.MaximumNArgs(1),
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		stringToLevel := map[string]zerolog.Level{
			"panic": zerolog.PanicLevel,
			"fatal": zerolog.FatalLevel,
			"error": zerolog.ErrorLevel,
			"warn":  zerolog.WarnLevel,
			"info":  zerolog.InfoLevel,
			"debug": zerolog.DebugLevel,
			"trace": zerolog.TraceLevel,
		}
		levelString := strings.TrimSpace(strings.ToLower(viper.GetString("logging-level")))
		if len(levelString) == 0 {
			levelString = "error"
		}
		zerolog.SetGlobalLevel(stringToLevel[levelString])
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.StampMilli})
		return nil
	}
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf("config file (default is $HOME/.%s.yaml)", rootCmd.Use))
	rootCmd.Flags().BoolP("verbose", "V", false, "Show verbose logging")
	viper.BindPFlag("verbose", rootCmd.Flags().Lookup("verbose"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".ldap-probe" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(fmt.Sprintf(".%s", rootCmd.Use))
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal().Err(err).Msg("Config file not found!")
	}
}
