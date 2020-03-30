//
// Author:: Salim Afiune Maya (<afiune@lacework.net>)
// Copyright:: Copyright 2020, Lacework Inc.
// License:: Apache License, Version 2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package api

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"
)

const defaultTimeout = 10 * time.Second

type Client struct {
	account    string
	apiVersion string
	logLevel   string
	baseURL    *url.URL
	auth       *authConfig
	c          *http.Client
	log        *zap.Logger

	Integrations *IntegrationsService
}

type Option interface {
	apply(c *Client) error
}

type clientFunc func(c *Client) error

func (fn clientFunc) apply(c *Client) error {
	return fn(c)
}

// New generates a new Lacework API client
//
// Example of basic usage
//
//   lacework, err := api.NewClient("demo")
//   if err == nil {
//       lacework.Integrations.List()
//   }
func NewClient(account string, opts ...Option) (*Client, error) {
	baseURL, err := url.Parse(fmt.Sprintf("https://%s.lacework.net", account))
	if err != nil {
		return nil, err
	}

	c := &Client{
		account:    account,
		baseURL:    baseURL,
		apiVersion: "v1",
		logLevel:   "info",
		auth: &authConfig{
			expiration: defaultTokenExpiryTime,
		},
		c: &http.Client{Timeout: defaultTimeout},
	}
	c.Integrations = &IntegrationsService{c}

	for _, opt := range opts {
		if err := opt.apply(c); err != nil {
			return c, err
		}
	}

	// if after applying all the custom options we don't have
	// a registered logger, initialize one
	if c.log == nil {
		c.initializeLogger()
	}

	c.log.Debug("api client created",
		zap.String("url", c.baseURL.String()),
		zap.String("log_level", c.logLevel),
		zap.Int("timeout", c.auth.expiration),
	)
	return c, nil
}

// WithURL sets the base URL, this options is only available for test purposes
func WithURL(baseURL string) Option {
	return clientFunc(func(c *Client) error {
		u, err := url.Parse(baseURL)
		if err != nil {
			return err
		}

		c.baseURL = u
		return nil
	})
}

// URL returns the base url configured
func (c *Client) URL() string {
	return c.baseURL.String()
}
