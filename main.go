package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"

	"github.com/ChimeraCoder/anaconda"
	"github.com/Rompei/inco"

	"gopkg.in/yaml.v2"
)

// TwitterInfo is information for Twitter API.
type TwitterInfo struct {
	WebhookURL        string `yaml:"webhookUrl"`
	ConsumerKey       string `yaml:"consumerKey"`
	ConsumerSecret    string `yaml:"consumerSecret"`
	AccessToken       string `yaml:"accessToken"`
	AccessTokenSecret string `yaml:"accessTokenSecret"`
}

// Valid validates TwitterInfo.
func (info TwitterInfo) Valid() bool {
	return info.WebhookURL != "" && info.ConsumerKey != "" && info.ConsumerSecret != "" && info.AccessToken != "" && info.AccessTokenSecret != ""
}

func main() {
	var settingFile string
	flag.StringVar(&settingFile, "s", "", "Path to setting file.")
	flag.Parse()
	if _, err := os.Stat(settingFile); err != nil {
		log.Fatalf("Setting file is not found in %s\n", settingFile)
	}

	b, err := ioutil.ReadFile(settingFile)
	if err != nil {
		log.Fatalf("Cant read file: %s\n", settingFile)
	}

	var info TwitterInfo

	if err = yaml.Unmarshal(b, &info); err != nil {
		log.Fatalf("Cant unmarshal yaml file: %s\n", settingFile)
	}

	if !info.Valid() {
		log.Fatal("Twitter information is not enough.")
	}

	anaconda.SetConsumerKey(info.ConsumerKey)
	anaconda.SetConsumerSecret(info.ConsumerSecret)
	api := anaconda.NewTwitterApi(info.AccessToken, info.AccessTokenSecret)

	api.SetLogger(anaconda.BasicLogger)
	stream := api.UserStream(url.Values{})
	for {
		select {
		case item := <-stream.C:
			switch status := item.(type) {
			case anaconda.DirectMessage:
				text := fmt.Sprintf("%s says %s", status.Sender.ScreenName, status.Text)
				msg := inco.Message{
					Username: "DM-notification",
					Text:     text,
				}
				if err = inco.Incoming(info.WebhookURL, &msg); err != nil {
					log.Fatal(err)
				}
			default:
			}
		}
	}
}
