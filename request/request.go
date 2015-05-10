package request

import (
	"net/http"
	"os"

	"github.com/smartystreets/go-aws-auth"
)

type Bucket struct {
	Name           string
	MaxKeys        uint
	Prefix         string
	Marker         string
	Contents       []BucketItem
	CommonPrefixes string
}

type BucketItem struct {
	Key          string
	LastModified string
	ETag         string
	Size         string
	Owner        BucketOwner
	StorageClass string
}

type BucketOwner struct {
	ID          string
	DisplayName string
}

var ACCESS_KEY, SECRET_KEY string

func init() {
	ACCESS_KEY = os.Getenv("AWS_ACCESS_KEY_ID")
	if ACCESS_KEY == "" {
		ACCESS_KEY = os.Getenv("AWS_ACCESS_KEY")
	}

	SECRET_KEY := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if SECRET_KEY == "" {
		SECRET_KEY = os.Getenv("AWS_SECRET_KEY")
	}
}

func signRequest(req *http.Request) *http.Request {
	return awsauth.Sign4(
		req,
		awsauth.Credentials{
			AccessKeyID:     ACCESS_KEY,
			SecretAccessKey: SECRET_KEY,
		},
	)
}

func Request(client *http.Client, req *http.Request) (*http.Response, error) {
	return client.Do(signRequest(req))
}
