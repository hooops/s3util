package s3url

import (
	"fmt"
	"net/url"
	"strings"
)

type S3URL struct {
	Bucket    string
	Path      string
	Region    string
	AccessKey string
	SecretKey string
}

func ParseS3URL(s3url string) (S3URL, error) {
	u, err := url.Parse(s3url)
	if err != nil {
		return S3URL{}, err
	}

	if u.Scheme != "s3" {
		return S3URL{}, fmt.Errorf("Not a s3:// url")
	}

	var user, pw string

	if u.User != nil {
		user = u.User.Username()
		pw, _ = u.User.Password()
	}

	s3 := S3URL{
		AccessKey: user,
		SecretKey: pw,
	}

	if s3.AccessKey != "" && s3.SecretKey == "" {
		return s3, fmt.Errorf("Both username and password must be set")
	}

	if s3.SecretKey != "" && s3.AccessKey == "" {
		return s3, fmt.Errorf("Both username and password must be set")
	}

	if strings.Contains(u.Host, ".") {
		strs := strings.Split(u.Host, ".")
		s3.Region = strs[len(strs)-1]
		if len(strs) == 2 {
			s3.Bucket = strs[0]
		} else {
			s3.Bucket = strings.Join(strs[:len(strs)-1], ".")
		}
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
