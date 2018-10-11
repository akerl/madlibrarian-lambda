package main

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/akerl/github-auth-lambda/session"
	"github.com/akerl/go-lambda/apigw/events"
	"github.com/akerl/go-lambda/mux"
	"github.com/akerl/go-lambda/mux/receivers/slack"
	"github.com/akerl/go-lambda/s3"
	"github.com/akerl/madlibrarian/utils"
	slackApi "github.com/nlopes/slack"
)

type cachedStory struct {
	Timestamp int64
	Story     *utils.Story
}

type storySet map[string]*cachedStory
type bucketSet map[string]storySet

var cache = make(bucketSet)
var sm *session.Manager

func cacheStory(bucketName, storyName string) (*utils.Story, error) {
	if cache[bucketName] == nil {
		cache[bucketName] = make(storySet)
	}
	cs := cache[bucketName][storyName]
	if cs == nil || cs.Timestamp+config.RefreshRate < time.Now().Unix() {
		storyObject := fmt.Sprintf("meta/%s.yml", storyName)

		storyObj, err := s3.GetObject(bucketName, storyObject)
		if err != nil {
			return &utils.Story{}, fmt.Errorf("config not found")
		}

		story, err := utils.NewStoryFromText(storyObj)
		if err != nil {
			return &utils.Story{}, fmt.Errorf("failed to parse config")
		}
		cache[bucketName][storyName] = &cachedStory{
			Timestamp: time.Now().Unix(),
			Story:     &story,
		}
	}
	return cache[bucketName][storyName].Story, nil
}

func parseStory(req events.Request) (string, string, error) {
	params := events.Params{Request: &req}
	bucketName := params.Lookup("bucket")
	storyName := params.Lookup("story")

	if bucketName == "" || storyName == "" {
		return "", "", fmt.Errorf("settings not provided")
	}
	return bucketName, storyName, nil
}

func aclCheck(aclName string, sess session.Session) bool {
	acl, ok := config.ACLs[aclName]
	if !ok {
		acl = config.ACLs["default"]
	}

	for _, aclEntry := range acl {
		if aclEntry == "anonymous" {
			return true
		}
		aclSlice := strings.SplitN(aclEntry, "/", 2)
		userOrgTeams, ok := sess.Memberships[aclSlice[0]]
		if ok {
			if len(aclSlice) == 1 {
				return true
			}
			for _, userTeam := range userOrgTeams {
				if userTeam == aclSlice[1] {
					return true
				}
			}
		}
	}
	return false
}

func authFunc(req events.Request) (events.Response, error) {
	bucketName, storyName, err := parseStory(req)
	if err != nil {
		return events.Response{
			StatusCode: 500,
			Body:       "failed to authenticate request",
		}, nil
	}

	sess, err := sm.Read(req)
	if err != nil {
		return events.Response{
			StatusCode: 500,
			Body:       "failed to authenticate request",
		}, nil
	}

	aclName := fmt.Sprintf("%s/%s", bucketName, storyName)
	if aclCheck(aclName, sess) {
		return events.Response{}, nil
	}

	if sess.Login == "" {
		authURL, err := url.Parse(config.AuthURL)
		if err != nil {
			return events.Response{}, err
		}

		returnURL := url.URL{
			Host:   req.Headers["Host"],
			Path:   req.Path,
			Scheme: "https",
		}
		returnValues := authURL.Query()
		returnValues.Set("redirect", returnURL.String())
		authURL.RawQuery = returnValues.Encode()

		return events.Response{
			StatusCode: 303,
			Headers: map[string]string{
				"Location": authURL.String(),
			},
		}, nil
	}

	return events.Response{}, nil
}

func loadQuote(req events.Request) (string, error) {
	bucketName, storyName, err := parseStory(req)
	if err != nil {
		return "", err
	}

	story, err := cacheStory(bucketName, storyName)
	if err != nil {
		return "", err
	}

	quote, err := story.Generate()
	if err != nil {
		return "", fmt.Errorf("failed to generate quote")
	}

	return quote, nil
}

func loadTextQuote(req events.Request) (events.Response, error) {
	body, err := loadQuote(req)
	if err != nil {
		return events.Response{}, err
	}
	return events.Succeed(body)
}

func loadSlackQuote(req events.Request) (*slackApi.Msg, error) {
	body, err := loadQuote(req)
	if err != nil {
		return &slackApi.Msg{}, err
	}
	return &slackApi.Msg{
		Text:         body,
		ResponseType: "in_channel",
	}, nil
}

func main() {
	var err error
	config, err = loadConfig()
	if err != nil {
		panic(err)
	}

	sm = &session.Manager{
		Name:     "session",
		SignKey:  config.SignKey,
		EncKey:   config.EncKey,
		Lifetime: config.Lifetime,
		Domain:   config.Domain,
	}

	d := mux.NewDispatcher(
		&slack.Handler{
			HandleFunc:  loadSlackQuote,
			SlackTokens: config.SlackTokens,
		},
		&mux.SimpleReceiver{
			HandleFunc: loadTextQuote,
			AuthFunc:   authFunc,
		},
	)
	mux.Start(d)
}
