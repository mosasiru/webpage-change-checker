package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/PuerkitoBio/goquery"
	"github.com/sergi/go-diff/diffmatchpatch"
)

type Config struct {
	Checker struct {
		CacheFile string       `toml:"cache_file"`
		Pages     []PageConfig `toml:"pages"`
	} `toml:"checker"`
	Slack SlackConfig `toml:"slack"`
}

type PageConfig struct {
	// required
	Name string `toml:"name"`
	// required
	URL string `toml:"url"`

	// CSS Selector (GoQuery)
	Selector string `toml:"selector"`

	// Default: 10 (second)
	Timeout int `toml:"timeout"`
	// Default: 600 (second)
	Interval int `toml:"interval"`

	NotifyNoChange bool `toml:"notify_no_change"`
	NotifyError    bool `toml:"notify_error"`
	// enum("slack")
	Notifier string `toml:"notifier"`
}

type SlackConfig struct {
	// required
	WebhookURL string `toml:"webhook_url"`

	Channel     string `toml:"channel"`
	UserName    string `toml:"username"`
	IconEmoji   string `toml:"icon_emoji"`
	IconURL     string `toml:"icon_url"`
	AlertPrefix string `toml:"alert_prefix"`
}

const (
	DefaultCacheFile = ".checker."
	DefaultTimeout   = 10
	DefaultInterval  = 600
)

var configFile = flag.String("c", "config.toml", "configuration file")

func main() {
	flag.Parse()

	var config Config
	_, err := toml.DecodeFile(*configFile, &config)
	if err != nil {
		panic(err)
	}
	log.Printf("load config: %#v", config)

	cacheFile := config.Checker.CacheFile
	if cacheFile == "" {
		cacheFile = DefaultCacheFile
	}
	for _, pc := range config.Checker.Pages {
		go func(pc PageConfig) {
			for {
				log.Printf("start: %s", pc.Name)
				diff, err := checkDiff(cacheFile+pc.Name, pc)
				if pc.Notifier == "slack" { // TODO clean
					if err != nil && pc.NotifyError {
						sa := SlackAttachment{
							Title:     pc.Name,
							TitleLink: pc.URL,
							PreText:   fmt.Sprintf("%s error", config.Slack.AlertPrefix),
							Text:      err.Error(),
							Color:     "danger",
						}
						postSlack(sa, config.Slack)
					} else if diff != "" {
						sa := SlackAttachment{
							Title:     pc.Name,
							TitleLink: pc.URL,
							PreText:   fmt.Sprintf("%s %s changed!", config.Slack.AlertPrefix, pc.Name),
							Text:      diff,
							Color:     "warning",
						}
						postSlack(sa, config.Slack)
					} else if diff == "" && pc.NotifyNoChange {
						sa := SlackAttachment{
							Title:     pc.Name,
							TitleLink: pc.URL,
							Text:      fmt.Sprintf("no change"),
							Color:     "good",
						}
						postSlack(sa, config.Slack)
					}
				}

				interval := pc.Interval
				if interval == 0 {
					interval = DefaultInterval
				}
				time.Sleep(time.Duration(interval) * time.Second)
			}
		}(pc)
	}
	for {
		time.Sleep(time.Minute)
	}
}

func checkDiff(cacheFile string, pc PageConfig) (string, error) {
	timeout := pc.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	res, err := client.Get(pc.URL)
	if err != nil {
		log.Printf("req error: %s, url: %s", err, pc.URL)
		return "", err
	}
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		log.Printf("goquery doc error: %s, url: %s", err, pc.URL)
		return "", err
	}

	var html string
	if pc.Selector == "" {
		html, err = doc.Html()
	} else {
		html, err = doc.Find(pc.Selector).Html()
	}
	if err != nil {
		log.Printf("HTML parse error: %s, url: %s, selector: %s", err, pc.URL, pc.Selector)
		return "", err
	}
	body := []byte(html)

	var diff string
	if preBody, err := ioutil.ReadFile(cacheFile); err == nil {
		if string(body) != string(preBody) {
			diff = buildDiffText(string(preBody), string(body))
		}
	}
	if err := ioutil.WriteFile(cacheFile, body, 0644); err != nil {
		log.Printf("write error: %s, %s", err, cacheFile)
		return "", err
	}
	return diff, nil
}

func buildDiffText(textA, textB string) string {
	dmp := diffmatchpatch.New()
	a, b, c := dmp.DiffLinesToChars(textA, textB)
	diffs := dmp.DiffCharsToLines(dmp.DiffMain(a, b, false), c)
	text := ""
	for _, diff := range diffs {
		if diff.Type == diffmatchpatch.DiffDelete {
			text += fmt.Sprintf("- %s\n", diff.Text)
		} else if diff.Type == diffmatchpatch.DiffInsert {
			text += fmt.Sprintf("+ %s\n", diff.Text)
		}
	}
	return text
}
