package main

import (
	"fmt"
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

var (
	ACCESS_KEY = os.Getenv("AWS_ACCESS_KEY_ID")
	SECRET_KEY = os.Getenv("AWS_SECRET_ACCESS_KEY")
)

func signRequest(req *http.Request) *http.Request {
	return awsauth.Sign4(
		req,
		awsauth.Credentials{
			AccessKeyID:     ACCESS_KEY,
			SecretAccessKey: SECRET_KEY,
		},
	)
}

func ErrExit(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

func Request(client *http.Client, req *http.Request) (*http.Response, error) {
	return client.Do(signRequest(req))
}
