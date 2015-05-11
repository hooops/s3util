package s3url

import (
	"fmt"
	"net/url"
	"strings"
)

type S3URL struct {
	Bucket string
	Path   string
	Region string
}

func ParseS3URL(s3url string) (S3URL, error) {
	u, err := url.Parse(s3url)
	if err != nil {
		return S3URL{}, err
	}

	if u.Scheme != "s3" {
		return S3URL{}, fmt.Errorf("Not a s3:// url")
	}

	s3 := S3URL{}

	if strings.Contains(u.Host, ".") {
		strs := strings.SplitN(u.Host, ".", 2)
		s3.Bucket = strs[0]
		s3.Region = strs[1]
	} else {
		s3.Bucket = u.Host
	}

	if u.Path != "" {
		if u.Path == "/" {
			s3.Path = ""
		} else {
			s3.Path = u.Path[1:]
		}
	}

	return s3, nil
}
