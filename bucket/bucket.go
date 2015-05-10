package bucket

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/erikh/s3util/common"
	"github.com/erikh/s3util/request"
	"github.com/erikh/s3util/s3url"
)

type bucket struct {
	Name     string
	MaxKeys  uint
	Prefix   string
	Marker   string
	Contents []struct {
		Key          string
		LastModified string
		ETag         string
		Size         string
		Owner        struct {
			ID          string
			DisplayName string
		}
		StorageClass string
	}
	CommonPrefixes string
}

type BucketClient struct {
	s3url  s3url.S3URL
	client request.Client
}

func NewBucketClient(s3url s3url.S3URL, c request.Client) *BucketClient {
	return &BucketClient{
		client: c,
		s3url:  s3url,
	}
}

func (bc *BucketClient) Find(found func(*string) error) error {
	marker := ""

	for {
		var bucket bucket

		req, err := http.NewRequest("GET", fmt.Sprintf("https://%s.%s?marker=%s&prefix=%s", bc.s3url.Bucket, bc.client.Host, url.QueryEscape(marker), url.QueryEscape(bc.s3url.Path)), nil)
		if err != nil {
			common.ErrExit("Could not complete request: %v", err)
		}

		resp, err := bc.client.Do(req)
		if err != nil {
			return fmt.Errorf("Could not complete request: %v", err)
		}

		if resp.StatusCode != 200 {
			fmt.Println("Ensure your region settings are correct.")
			return fmt.Errorf("Could not read bucket: fatal error.")
		}

		content, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("Failure during download: %v", err)
		}

		if err := xml.Unmarshal(content, &bucket); err != nil {
			return fmt.Errorf("Failure during XML parse: %v", err)
		}

		if len(bucket.Contents) == 0 {
			break
		}

		for _, item := range bucket.Contents {
			str := item.Key
			if err := found(&str); err != nil {
				return err
			}
		}

		marker = bucket.Contents[len(bucket.Contents)-1].Key
	}

	return nil
}
