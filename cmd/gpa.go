package cmd

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Delta456/box-cli-maker/v2"
	"github.com/PuerkitoBio/goquery"
	"github.com/briandowns/spinner"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// gpaCmd represents the gpa command
var gpaCmd = &cobra.Command{
	Use:   "gpa",
	Short: "Get your GPA (weighted and unweighted) as well as your rank",
	Run: func(cmd *cobra.Command, args []string) {
		if viper.GetString("username") == "" || viper.GetString("password") == "" || viper.GetString("url") == "" {
			color.Red.Println("❌ Please define a username, password, and home access url in $HOME/.hac.yaml! (or, use hac setup to run a wizard that does it for you)")
			os.Exit(1)
		}

		sp := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		sp.Suffix = color.Blue.Render(" Fetching data from HAC...")

		sp.Start()
		cj, err := cookiejar.New(nil)
		cobra.CheckErr(err)
		client := &http.Client{Jar: cj}
		rvtReq, err := client.Get(fmt.Sprintf("%v/HomeAccess/Account/LogOn", viper.GetString("url")))
		cobra.CheckErr(err)
		defer rvtReq.Body.Close()

		rvtDoc, err := goquery.NewDocumentFromReader(rvtReq.Body)
		cobra.CheckErr(err)

		sel := rvtDoc.Find("input[name='__RequestVerificationToken']")
		rvt, ex := sel.Attr("value")
		if !ex {
			color.Red.Println("❌ Failed to fetch request token from HAC!")
			os.Exit(1)
		}

		reqBody := url.Values{}
		reqBody.Add("__RequestVerificationToken", rvt)
		reqBody.Add("SCKTY00328510CustomEnabled", "False")
		reqBody.Add("SCKTY00436568CustomEnabled", "False")
		reqBody.Add("Database", "10")
		reqBody.Add("VerificationOption", "UsernamePassword")
		reqBody.Add("LogOnDetails.UserName", viper.GetString("username"))
		reqBody.Add("LogOnDetails.Password", viper.GetString("password"))
		reqBody.Add("tempUN", "")
		reqBody.Add("tempPW", "")

		loginResp, err := client.PostForm(fmt.Sprintf("%v/HomeAccess/Account/LogOn", viper.GetString("url")), reqBody)
		client.Jar.SetCookies(loginResp.Request.URL, loginResp.Cookies())
		cobra.CheckErr(err)

		transcriptPage, err := client.Get(fmt.Sprintf("%v/HomeAccess/Content/Student/Transcript.aspx", viper.GetString("url")))
		cobra.CheckErr(err)

		defer transcriptPage.Body.Close()

		transcriptDoc, err := goquery.NewDocumentFromReader(transcriptPage.Body)
		cobra.CheckErr(err)

		wGPA := transcriptDoc.Find("#plnMain_rpTranscriptGroup_lblGPACum1").Text()
		uwGPA := transcriptDoc.Find("#plnMain_rpTranscriptGroup_lblGPACum2").Text()
		rank := transcriptDoc.Find("#plnMain_rpTranscriptGroup_lblGPARank1").Text()
		sp.Stop()
		bx := box.New(box.Config{Px: 0, Py: 0, Type: "Round", Color: "Cyan"})
		bx.Println(
			"Rank and GPA",
			fmt.Sprintf(
				"%v %v\n%v %v\n%v %v",
				color.Bold.Render("Weighted GPA: "),
				strings.TrimSpace(wGPA),
				color.Bold.Render("Unweighted GPA: "),
				strings.TrimSpace(uwGPA),
				color.Bold.Render("Rank: "),
				strings.TrimSpace(rank),
			),
		)
	},
}

func init() {
	rootCmd.AddCommand(gpaCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// gpaCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// gpaCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
