package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"unicode/utf8"

	"./config"
	m "github.com/kaias1jp/go-mastodon"
)

func main() {
	confDir := "./env/"
	appMode := os.Getenv("APP_MODE") // 動作環境をAPP_MODEという形で環境変数に格納する
	if appMode == "" {
		panic("failed to get application mode, check whether APP_MODE is set.")
	}

	conf, err := config.NewConfig(confDir, appMode) // 引数に渡す
	if err != nil {
		panic(err.Error())
	}
	tagname := conf.APP.Tagname
	config := &m.Config{
		Server:       conf.APP.Server,
		ClientID:     conf.APP.ClientID,
		ClientSecret: conf.APP.ClientSecret,
	}

	c := m.NewClient(config)

	err = c.Authenticate(context.Background(), conf.APP.Email, conf.APP.Password)
	if err != nil {
		log.Fatal(err)
	}

	wsc := c.NewWSClient()

	q, err := wsc.StreamingWSHashtag(context.Background(), tagname, true)
	if err != nil {
		fmt.Printf("  ERR: %s\n", err)
	}

	for e := range q {
		if t, ok := e.(*m.UpdateEvent); ok {
			var splited []string
			var content string
			var keyword []string
			var lang string = "ja"
			var mode string = "title"
			var op string = "and"
			var mention string = ""

			content = removeTag(t.Status.Content)
			splited = strings.Split(changeZenspace(content), " ")
			if len(splited) >= 1 {
				for _, k := range splited {
					switch k {
					case "-en":
						lang = "en"
					case "-all":
						mode = ""
					case "-or":
						op = "or"
					case "#" + tagname:

					case "":

					default:
						if strings.Contains(k, "@") {
							mention = mention + " " + k
						} else {
							keyword = append(keyword, k)
						}
					}
				}
			}
			fmt.Printf("\x1b[37m[%s] \x1b[35m%-20s: \x1b[33m%s\n", t.Status.CreatedAt.Local().Format("15:04:05"), t.Status.Account.Acct, content)
			a, err := c.GetAccountCurrentUser(context.Background())
			if err != nil {
				fmt.Printf("  ERR: %s\n", err)
			}
			if t.Status.Account.ID != a.ID {
				if len(keyword) > 0 {
					client := &http.Client{}
					req, err := http.NewRequest("GET", "https://socialapi.app/api/googlebooksearch", nil)
					if err != nil {
						log.Print(err)
						os.Exit(1)
					}

					q := req.URL.Query()
					q.Add("lang", lang)
					q.Add("op", op)
					q.Add("mode", mode)
					q.Add("keyword", strings.Join(keyword, " "))
					q.Add("orderBy", "newest")
					req.URL.RawQuery = q.Encode()
					resp, err := client.Do(req)

					if err != nil {
						fmt.Println(err)
						fmt.Println("Errored when sending request to the server")
						continue
					}

					defer resp.Body.Close()
					resp_body, _ := ioutil.ReadAll(resp.Body)

					fmt.Println(resp.Status)
					result := string(resp_body)

					length := utf8.RuneCountInString(result)
					if length > 450 {
						length = 450
					}
					ac, _ := c.GetAccount(context.Background(), t.Status.Account.ID)
					c.PostStatus(context.Background(), &m.Toot{
						Status: "@" + ac.Acct + mention + " googlebooksAPIの結果です\n" +
							string([]rune(result)[:length]),
						Visibility:  "direct",
						InReplyToID: t.Status.ID,
					})
				}
			}
		}
	}
}

func arrayContains(arr []m.Mention, id m.ID) bool {
	for _, v := range arr {
		if v.ID == id {
			return true
		}
	}
	return false
}

func removeTag(str string) string {
	rep := regexp.MustCompile(`<("[^"]*"|'[^']*'|[^'">])*>`)
	str = rep.ReplaceAllString(str, "")
	return str
}
func changeZenspace(str string) string {
	rep := regexp.MustCompile(`　`)
	str = rep.ReplaceAllString(str, " ")
	return str
}
