package cmd

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/Delta456/box-cli-maker/v2"
	"github.com/PuerkitoBio/goquery"
	"github.com/briandowns/spinner"
	"github.com/cheynewallace/tabby"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// overviewCmd represents the overview command
var overviewCmd = &cobra.Command{
	Use:   "overview",
	Short: "Overview of grades",
	Long:  `Get an overview of curent grades`,
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
			color.Red.Println("❌ Failed to fetch Request Token from HAC!")
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

		assignmentsResp, err := client.Get(fmt.Sprintf("%v/HomeAccess/Content/Student/Assignments.aspx", viper.GetString("url")))
		client.Jar.SetCookies(assignmentsResp.Request.URL, assignmentsResp.Cookies())
		cobra.CheckErr(err)

		defer assignmentsResp.Body.Close()

		assignments, err := goquery.NewDocumentFromReader(assignmentsResp.Body)
		cobra.CheckErr(err)
		sp.Stop()

		buf := new(bytes.Buffer)
		wtr := tabwriter.NewWriter(buf, 0, 0, 2, ' ', 0)

		t := tabby.NewCustom(wtr)

		t.AddHeader("Class ID", "Name", "Average")

		assignments.Find(".AssignmentClass").Each(func(i int, s *goquery.Selection) {
			name := s.Find("a[class='sg-header-heading']").Text()
			avg := s.Find(fmt.Sprintf("span[id='plnMain_rptAssigmnetsByCourse_lblHdrAverage_%v']", i)).Text()
			t.AddLine(i, strings.TrimSpace(strings.TrimSpace(name)[9:]), colorGrade(strings.TrimSpace(avg[3:])))
		})

		t.Print()
		bx := box.New(box.Config{Px: 0, Py: 0, Type: "Round", Color: "Cyan"})
		bx.Println("", buf.String())

	},
}

func colorGrade(avg string) string {
	numAvg, _ := strconv.Atoi(avg)
	switch true {
	case numAvg >= 90:
		return color.Green.Render(avg)
	case numAvg >= 80:
		return color.Blue.Render(avg)
	case numAvg >= 70:
		return color.Yellow.Render(avg)
	default:
		return color.Red.Render(avg)
	}

}

func init() {
	rootCmd.AddCommand(overviewCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// overviewCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// overviewCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
