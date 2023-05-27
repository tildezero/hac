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

// classCmd represents the class command
var classCmd = &cobra.Command{
	Use:   "class [id]",
	Short: "Get grades for a specifed class",
	Long: `Get a full list of grades for a class based on its ID.

	To get the ID for a class, run hac overview, and use the ID on the left column, and use that as an argument to hac class.

	For example, if hac overview says that AP Calculus AB has an id of 1, to view a full set of grades for that class, one would run hac class 1`,
	Run: func(cmd *cobra.Command, args []string) {
		classID, err := strconv.Atoi(args[0])
		if err != nil {
			color.Red.Println("❌ Class ID must be a number!")
			os.Exit(1)
		}

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

		t.AddHeader("Assigment", "Category", "Grade", "Due", "Assigned")

		classHTML := assignments.Find(".AssignmentClass").Nodes[classID]
		class := goquery.NewDocumentFromNode(classHTML)
		class.Find("div[class='sg-content-grid'] tr[class='sg-asp-table-data-row']").Each(func(i int, s *goquery.Selection) {
			tds := s.Find("td")
			dueDate := strings.TrimSpace(tds.Eq(0).Text())
			assignedDate := strings.TrimSpace(tds.Eq(1).Text())
			category := strings.TrimSpace(tds.Eq(3).Text())
			score := strings.TrimSpace(tds.Eq(4).Text())
			total := strings.TrimSpace(tds.Eq(5).Text())
			name := strings.TrimSpace(s.Find("a").Text())

			if name != "" {
				t.AddLine(name, category, fmt.Sprintf("%v/%v", score, total), dueDate, assignedDate)
			}
		})

		t.Print()
		bx := box.New(box.Config{Px: 0, Py: 0, Type: "Round", Color: "Cyan"})
		className := strings.TrimSpace(
			strings.TrimSpace(
				class.Find("a[class='sg-header-heading']").Text(),
			)[9:],
		)
		classAvg := strings.TrimSpace(
			class.Find(
				fmt.Sprintf("span[id='plnMain_rptAssigmnetsByCourse_lblHdrAverage_%v']", classID),
			).Text(),
		)[3:]
		bx.Println(fmt.Sprintf("Grades for %v - Average %v", className, classAvg), buf.String())

	},
}

func init() {
	rootCmd.AddCommand(classCmd)
}
