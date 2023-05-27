package cmd

import (
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile  string
	username string
	password string
	hacUrl   string

	rootCmd = &cobra.Command{
		Use:   "hac",
		Short: "A CLI for Home Access Center",
		Long: `Home Access Center is a software used by many school districts to display a student's grades and other vital information
Using this program, students can quickly check their grades, and other information directly from the terminal
`,
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.hac.yaml)")
	rootCmd.PersistentFlags().StringVarP(&username, "username", "n", "", "home access username")

	rootCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "home access password")

	rootCmd.PersistentFlags().StringVarP(&hacUrl, "url", "u", "", "home access url")

	viper.BindPFlag("username", rootCmd.PersistentFlags().Lookup("username"))

	viper.BindPFlag("password", rootCmd.PersistentFlags().Lookup("password"))

	viper.BindPFlag("url", rootCmd.PersistentFlags().Lookup("url"))

}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".hac")
		viper.SafeWriteConfigAs(path.Join(home, ".hac.yaml"))
	}

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	cobra.CheckErr(err)
}
