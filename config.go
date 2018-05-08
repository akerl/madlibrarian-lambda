package main

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/akerl/go-lambda/s3"
	"github.com/ghodss/yaml"
)

type configFile struct {
	Lifetime      int                 `json:"lifetime"`
	Domain        string              `json:"domain"`
	Base64SignKey string              `json:"signkey"`
	Base64EncKey  string              `json:"enckey"`
	SignKey       []byte              `json:"-"`
	EncKey        []byte              `json:"-"`
	SlackTokens   []string            `json:"slacktokens"`
	ACLS          map[string][]string `json:"acls"`
	AuthURL       string              `json:"acls"`
}

var config *configFile

func loadConfig() (*configFile, error) {
	c := configFile{}

	bucket := os.Getenv("S3_BUCKET")
	path := os.Getenv("S3_KEY")
	if bucket == "" || path == "" {
		return &c, fmt.Errorf("config location not provided")
	}

	obj, err := s3.GetObject(bucket, path)
	if err != nil {
		return &c, err
	}

	err = yaml.Unmarshal(obj, &c)
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
