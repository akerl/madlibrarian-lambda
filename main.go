package main

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/akerl/github-auth-lambda/session"
	"github.com/akerl/go-lambda/apigw/dispatch"
	"github.com/akerl/go-lambda/apigw/dispatch/handlers/slack"
	"github.com/akerl/go-lambda/apigw/dispatch/handlers/text"
	"github.com/akerl/go-lambda/apigw/events"
	"github.com/akerl/go-lambda/s3"
	"github.com/akerl/madlibrarian/utils"
	slackApi "github.com/nlopes/slack"
)

type storySet map[string]*utils.Story
type bucketSet map[string]storySet

var cache = make(bucketSet)
var sm *session.Manager

func cacheStory(bucketName, storyName string) (*utils.Story, error) {
	if cache[bucketName] == nil {
		cache[bucketName] = make(storySet)
	}
	if cache[bucketName][storyName] == nil {
		storyObject := fmt.Sprintf("meta/%s.yml", storyName)

		config, err := s3.GetObject(bucketName, storyObject)
		if err != nil {
			return &utils.Story{}, fmt.Errorf("config not found")
		}

		story, err := utils.NewStoryFromText(config)
		if err != nil {
			return &utils.Story{}, fmt.Errorf("failed to parse config")
		}
		cache[bucketName][storyName] = &story
	}
	return cache[bucketName][storyName], nil
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
		returnURL := url.URL{
			Host:   req.Headers["Host"],
			Path:   req.Path,
			Scheme: "https",
		}
		sess.Target = returnURL.String()

		return events.Response{
			StatusCode: 303,
			Headers: map[string]string{
				"Location": config.AuthURL,
			},
		}, nil
	}

	return events.Response{
		StatusCode: 403,
		Body:       "not authenticated",
	}, nil
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

	d := dispatch.Dispatcher{
		Receivers: []dispatch.Receiver{
			&slack.Handler{
				Func:        loadSlackQuote,
				SlackTokens: config.SlackTokens,
			},
			&text.Handler{
				Func:     loadQuote,
				AuthFunc: authFunc,
			},
		},
	}
	d.Start()
}
