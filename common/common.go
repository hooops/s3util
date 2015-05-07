package common

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

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

type GetConfig struct {
	Client     *http.Client
	Pathchan   chan *string
	Donechan   chan struct{}
	BucketName string
	LocalPath  string
	Host       string
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

func (gc *GetConfig) Get() (bool, bool) {
	target := <-gc.Pathchan
	if target == nil {
		gc.Donechan <- struct{}{}
		return false, true
	}

	fullPath := filepath.Join(gc.LocalPath, *target)

	if strings.HasSuffix(*target, "/") {
		os.MkdirAll(fullPath, 0755)
		return true, false
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s.%s/%s", gc.BucketName, gc.Host, *target), nil)
	if err != nil {
		gc.Pathchan <- target
		return false, false
	}

	resp, err := Request(gc.Client, req)
	if err != nil {
		gc.Pathchan <- target
		return false, false
	}

	if resp.StatusCode != 200 {
		fmt.Printf("Received status %d downloading; cannot continue.\n", resp.StatusCode)
		os.Exit(1)
	}

	// if there's any error requesting the url, retry.  silently do so if we
	// get an EOF error during the request. This usually means we're doing too
	// many requests at once.  be loud if we get some other error.

	// FIXME There's a race here between deleted files happening after the list
	// is performed. It lives in the channel for a while on large lists, so we
	// might need to stat it or react to the errors better or something. The
	// fail state here is to retry infinitely.
	if _, ok := err.(*url.Error); ok {
		gc.Pathchan <- target
		return false, false
	} else if err != nil {
		fmt.Printf("Error: %v - retrying\n", err)
		gc.Pathchan <- target
		return false, false
	}

	if err := os.MkdirAll(filepath.Dir(fullPath), 0700); err != nil {
		ErrExit("Could not create directory for download path: %v", err)
	}

	f, err := os.Create(fullPath)
	if err != nil {
		ErrExit("Could not create file: %v", err)
	}

	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		gc.Pathchan <- target
		return false, false
	}

	fmt.Println(*target, "->", fullPath)

	return true, false
}
