package request

import (
	"fmt"
	"net/http"

	"github.com/smartystreets/go-aws-auth"
)

type Client struct {
	HTTP *http.Client
	AWS  awsauth.Credentials
	Host string
}

func NewClient(access, secret, host, region string) Client {
	c := Client{
		AWS: awsauth.Credentials{
			AccessKeyID:     access,
			SecretAccessKey: secret,
		},
		Host: host,
		HTTP: new(http.Client),
	}

	if c.Host == "" {
		if region != "" {
			c.Host = fmt.Sprintf("s3-%s.amazonaws.com", region)
		} else {
			c.Host = "s3.amazonaws.com"
		}
	}

	return c
}

func (c Client) Do(req *http.Request) (*http.Response, error) {
	return c.HTTP.Do(c.signClient(req))
}

func (c Client) signClient(req *http.Request) *http.Request {
	return awsauth.Sign4(req, c.AWS)
}
