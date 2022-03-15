package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"syscall"

	"golang.org/x/net/html"
	"golang.org/x/net/publicsuffix"
	"golang.org/x/term"
)

type opts struct {
	username  string
	password  string
	networkID string
	date      string
}

func (o *opts) validate() error {
	if len(o.username) == 0 {
		return errors.New("username required")
	}

	if len(o.password) == 0 {
		psswd, err := term.ReadPassword(syscall.Stdin)
		if err != nil {
			return err
		}
		o.password = string(psswd)
	}

	if len(o.date) == 0 {
		return errors.New("specify a data or date range")
	}
	r := regexp.MustCompile(`(?m)^\d{4}-\d\d-\d\d(?:to\d{4}-\d\d-\d\d)?$`)
	if !r.MatchString(o.date) {
		return fmt.Errorf("invalid date or date range: YYYY-MM-DD or YYYY-MM-DDtoYYYY-MM-DD")
	}
	return nil
}

func main() {

	o := &opts{}
	flag.StringVar(&o.username, "user", "", "OpenDNS email or username.")
	flag.StringVar(&o.password, "password", "", "OpenDNS user password.")
	flag.StringVar(&o.networkID, "network-id", "all", "Network ID.")
	flag.StringVar(&o.date, "date", "", "YYYY-MM-DD or YYYY-MM-DDtoYYYY-MM-DD for range")
	flag.Parse()

	err := o.validate()
	if err != nil {
		flag.Usage()
		log.Fatalf("Error parsing flags: %s", err)
	}

	err = run(o)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run(o *opts) error {

	// initialize http.Client that supports cookies
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return fmt.Errorf("unable to initialize cookie jar: %v", err)
	}
	client := &http.Client{Jar: jar}

	err = login(client, o.username, o.password)
	if err != nil {
		return err
	}

	return getStats(client, o.networkID, o.date)
}

func getStats(client *http.Client, networkID, date string) error {
	var done bool
	for page := 1; !done; page++ {

		// get page of csv records
		response, err := client.Get(fmt.Sprintf("https://dashboard.opendns.com/stats/%s/topdomains/%s/page%d.csv", networkID, date, page))
		if err != nil {
			return err
		}

		// if Content-Disposition is not set, error: date out of range or wrong network
		if len(response.Header.Get("Content-Disposition")) == 0 {
			return fmt.Errorf("invalid network and/or date range")
		}

		scanner := bufio.NewScanner(response.Body)

		// only print the header on the first page
		if scanner.Scan() && page == 1 {
			fmt.Println(scanner.Text())
		}
		var recordCount int
		for scanner.Scan() {
			recordCount++
			fmt.Println(scanner.Text())
		}
		// if there were less than 200 records, it was the last page
		done = recordCount != 200
	}
	return nil
}

func login(client *http.Client, username string, password string) error {
	token, err := getFormToken(client)
	if err != nil {
		return fmt.Errorf("unable to retreive formtoken: %s", err)
	}
	form := url.Values{}
	form.Add("formtoken", token)
	form.Add("username", username)
	form.Add("password", password)
	form.Add("sign_in_submit", "foo")
	response, err := client.PostForm("https://login.opendns.com", form)
	if err != nil {
		return fmt.Errorf("error logging in: %s", err)
	}
	for _, cookie := range response.Cookies() {
		if cookie.Name == "OPENDNS_ACCOUNT" {
			return nil
		}
	}
	return fmt.Errorf("authentication failed")
}

func getFormToken(client *http.Client) (string, error) {
	response, err := client.Get("https://login.opendns.com")
	if err != nil {
		return "", fmt.Errorf("unable to retrieve login form: %s", err)
	}
	doc, err := html.Parse(response.Body)
	if err != nil {
		return "", fmt.Errorf("unable to parse login form: %s", err)
	}
	token := findFormTokenValue(findFormTokenInput(findLoginForm(doc)))
	if len(token) == 0 {
		return "", fmt.Errorf("formtoken not found")
	}
	return token, nil
}

func findLoginForm(n *html.Node) *html.Node {
	if n == nil {
		return nil
	}
	if n.Type == html.ElementNode && n.Data == "form" {
		for _, attr := range n.Attr {
			if attr.Key == "name" && attr.Val == "signin" {
				return n
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		lf := findLoginForm(c)
		if lf != nil {
			return lf
		}
	}
	return nil
}

func findFormTokenInput(n *html.Node) *html.Node {
	if n == nil {
		return nil
	}
	if n.Type == html.ElementNode && n.Data == "input" {
		for _, attr := range n.Attr {
			if attr.Key == "name" && attr.Val == "formtoken" {
				return n
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		lf := findFormTokenInput(c)
		if lf != nil {
			return lf
		}
	}
	return nil
}

func findFormTokenValue(n *html.Node) string {
	if n == nil {
		return ""
	}
	for _, attr := range n.Attr {
		if attr.Key == "value" {
			return attr.Val
		}
	}
	return ""
}
