package s3url

import (
	"fmt"
	"net/url"
)

type S3URL struct {
	Bucket string
	Path   string
}

func ParseS3URL(s3url string) (S3URL, error) {
	u, err := url.Parse(s3url)
	if err != nil {
		return S3URL{}, err
	}

	if u.Scheme != "s3" {
		return S3URL{}, fmt.Errorf("Not a s3:// url")
	}

	return S3URL{
		Bucket: u.Host,
		Path:   u.Path,
	}, nil
}
