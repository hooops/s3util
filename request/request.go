package request

import (
	"net/http"

	"github.com/smartystreets/go-aws-auth"

	"github.com/erikh/s3util/env"
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

func signRequest(req *http.Request) *http.Request {
	return awsauth.Sign4(
		req,
		awsauth.Credentials{
			AccessKeyID:     env.ACCESS_KEY,
			SecretAccessKey: env.SECRET_KEY,
		},
	)
}

func Request(client *http.Client, req *http.Request) (*http.Response, error) {
	return client.Do(signRequest(req))
}
