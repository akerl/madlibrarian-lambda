package main

import (
	"encoding/base64"
	"fmt"

	"github.com/akerl/go-lambda/s3"
)

type configFile struct {
	Lifetime      int                 `json:"lifetime"`
	Domain        string              `json:"domain"`
	Base64SignKey string              `json:"signkey"`
	Base64EncKey  string              `json:"enckey"`
	SignKey       []byte              `json:"-"`
	EncKey        []byte              `json:"-"`
	SlackTokens   []string            `json:"slacktokens"`
	ACLs          map[string][]string `json:"acls"`
	AuthURL       string              `json:"authurl"`
}

var config *configFile

func loadConfig() (*configFile, error) {
	c := configFile{}
	err := s3.GetConfigFromEnv(&c)
	if err != nil {
		return &c, err
	}

	if c.Lifetime == 0 {
		c.Lifetime = 86400
	}

	if c.Base64SignKey == "" || c.Base64EncKey == "" {
		return &c, fmt.Errorf("signing and encryption keys not set")
	}

	c.SignKey, err = base64.URLEncoding.DecodeString(c.Base64SignKey)
	if err != nil {
		return &c, err
	}
	c.EncKey, err = base64.URLEncoding.DecodeString(c.Base64EncKey)
	if err != nil {
		return &c, err
	}

	if c.AuthURL == "" {
		return &c, fmt.Errorf("auth url not set")
	}

	return &c, nil
}
