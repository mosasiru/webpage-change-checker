package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
)

type SlackPayload struct {
	Channel     string            `json:"channel"`
	UserName    string            `json:"username"`
	IconEmoji   string            `json:"icon_emoji"`
	IconURL     string            `json:"icon_url"`
	Attathments []SlackAttachment `json:"attachments"`
}

type SlackAttachment struct {
	Title     string `json:"title"`
	TitleLink string `json:"title_link"`
	PreText   string `json:"pretext"`
	Text      string `json:"text"`
	Color     string `json:"color"`
}

func postSlack(sa SlackAttachment, sc SlackConfig) error {
	log.Printf("start to post slack: %#v", sa)
	p := SlackPayload{
		Channel:     sc.Channel,
		UserName:    sc.UserName,
		IconEmoji:   sc.IconEmoji,
		IconURL:     sc.IconURL,
		Attathments: []SlackAttachment{sa},
	}
	payload, err := json.Marshal(p)
	if err != nil {
		log.Printf("json marshal error: %s", err)
		return err
	}
	res, err := http.PostForm(sc.WebhookURL, url.Values{"payload": {string(payload)}})
	if err != nil {
		log.Printf("slack request error: %s", err)
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		log.Printf("cslack status error: %d", res.StatusCode)
		return errors.New(res.Status)
	}
	return nil
}
