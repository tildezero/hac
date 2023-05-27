package cmd

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"

	"github.com/AlecAivazis/survey/v2"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup the Home Access CLI configuration file",
	Long: `Adds relevant configuration settings to the HAC CLI config file for future use
The HAC CLI config file is located at $HOME/.hac.yaml and could also be manually configured with the relevant values`,
	Run: func(cmd *cobra.Command, args []string) {
		uhd, err := os.UserHomeDir()
		viper.SafeWriteConfigAs(path.Join(uhd, ".hac.yaml"))
		cobra.CheckErr(err)
		overwrite := false
		overwrite_question := &survey.Confirm{
			Message: fmt.Sprintf("This command will OVERWRITE your current configuration file (located at %v), is that OK?", path.Join(uhd, ".hac.yaml")),
		}
		if err = survey.AskOne(overwrite_question, &overwrite, survey.WithValidator(survey.Required)); !overwrite || err != nil {
			cobra.CheckErr(err)
			fmt.Println("Exiting...")
			os.Exit(1)
		}

		var hacUrl string
		urlQuestion := &survey.Input{
			Message: "Enter the URL to your Home Access Center page without the /HomeAccess and with the https://",
			Help:    "For example: https://accesscenter.roundrockisd.org would be correct (notice that there isn't a slash at the end, no HomeAccess, and it has https:// at the beginning)",
		}
		err = survey.AskOne(urlQuestion, &hacUrl, survey.WithValidator(survey.Required), survey.WithValidator(valUrl))
		cobra.CheckErr(err)

		viper.Set("url", hacUrl)

		var username string
		usernameQuestion := &survey.Input{
			Message: "Enter your Home Access Center username",
		}
		err = survey.AskOne(usernameQuestion, &username, survey.WithValidator(survey.Required))
		cobra.CheckErr(err)

		viper.Set("username", username)

		var password string
		passwordQuestion := &survey.Password{
			Message: "Enter your Home Access Center password",
		}
		err = survey.AskOne(passwordQuestion, &password, survey.WithValidator(survey.Required))
		cobra.CheckErr(err)

		viper.Set("password", password)

		err = viper.WriteConfigAs(path.Join(uhd, ".hac.yaml"))
		cobra.CheckErr(err)
		fmt.Printf("✅ %v %v\n", color.Success.Render("Wrote settings to"), color.Blue.Render(path.Join(uhd, ".hac.yaml")))
	},
}

func valUrl(val interface{}) error {
	if _, ok := val.(string); !ok {
		return errors.New("input must be a string")
	}
	_, err := url.ParseRequestURI(val.(string))
	if err != nil {
		return errors.New("input must be a URL")
	}
	return nil
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
