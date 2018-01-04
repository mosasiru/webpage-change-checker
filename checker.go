package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
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
	// Default: 10 (second)
	Interval int `toml:"interval"`

	NotifyNoChange bool   `toml:"notify_no_change"`
	NotifyError    bool   `toml:"notify_error"`
	// enum("slack")
	Notifier       string `toml:"notifier"`
}

type SlackConfig struct {
	// required
	WebhookURL string `toml:"webhook_url"`

	Channel    string `toml:"channel"`
	UserName   string `toml:"username"`
	IconEmoji  string `toml:"icon_emoji"`
	IconURL    string `toml:"icon_url"`
	PrefixText string `toml:"prefix_text"`
}

const (
	DefaultCacheFile = ".checker."
	DefaultTimeout   = 10
	DefaultInterval  = 10
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
				if pc.Notifier == "slack" { // TODO
					if err != nil && pc.NotifyError {
						postSlack(fmt.Sprintf("%s error: %s", pc.Name, err), config.Slack)
					}
					if diff != "" {
						postSlack(fmt.Sprintf("%s changed!\n %s", pc.Name, diff), config.Slack)
					}
					if diff == "" && pc.NotifyNoChange {
						postSlack(fmt.Sprintf("%s no change.", pc.Name), config.Slack)
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
			diff = buildDiffText(string(body), string(preBody))
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

func postSlack(text string, sc SlackConfig) error {
	text = sc.PrefixText + text
	log.Printf("start post slack: %s", text)

	var params = struct {
		Text      string `json:"text"`
		Channel   string `json:"channel"`
		UserName  string `json:"username"`
		IconEmoji string `json:"icon_emoji"`
		IconURL   string `json:"icon_url"`
	}{
		Text:      text,
		Channel:   sc.Channel,
		UserName:  sc.UserName,
		IconEmoji: sc.IconEmoji,
		IconURL:   sc.IconURL,
	}
	payload, err := json.Marshal(params)
	if err != nil {
		log.Printf("json marshal error: %s", err)
		return err
	}
	res, err := http.PostForm(sc.WebhookURL, url.Values{"payload": {string(payload)}})
	if err != nil {
		log.Printf("slack request error: %s", err)
		return err
	}
	if res.StatusCode != http.StatusOK {
		log.Printf("cslack status error: %d", res.StatusCode)
		return errors.New(res.Status)
	}
	return nil
}