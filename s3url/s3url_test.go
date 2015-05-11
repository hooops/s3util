package s3url

import "testing"

type truth struct {
	url        string
	succeeds   bool
	bucket     string
	path       string
	region     string
	access_key string
	secret_key string
}

var truthTable = []truth{
	truth{
		url:      "s3://test",
		succeeds: true,
		bucket:   "test",
	},
	truth{
		url:      "s3://",
		succeeds: false,
	},
	truth{
		url:      "s3://test/test",
		succeeds: true,
		bucket:   "test",
		path:     "test",
	},
	truth{
		url:      "s3://quux/test",
		succeeds: true,
		bucket:   "quux",
		path:     "test",
	},
	truth{
		url:      "s3://quux.us-west-1/test",
		succeeds: true,
		bucket:   "quux",
		region:   "us-west-1",
		path:     "test",
	},
	truth{
		url:      "s3://quux.bar.us-east-1/test-1",
		succeeds: true,
		bucket:   "quux.bar",
		region:   "us-east-1",
		path:     "test-1",
	},
	truth{
		url:      "s3://quux.bar.baz.us-east-1/test-1",
		succeeds: true,
		bucket:   "quux.bar.baz",
		region:   "us-east-1",
		path:     "test-1",
	},
	truth{
		url:        "s3://akia:baz@quux.bar.us-east-1/test-1",
		succeeds:   true,
		bucket:     "quux.bar",
		region:     "us-east-1",
		path:       "test-1",
		access_key: "akia",
		secret_key: "baz",
	},
	truth{
		url:      "s3://akia@quux.bar.us-east-1/test-1",
		succeeds: false,
	},
}

func TestBasic(t *testing.T) {
	for _, truth := range truthTable {
		s3, err := ParseS3URL(truth.url)
		if err != nil {
			if truth.succeeds {
				t.Fatal(err)
			} else {
				continue
			}
		}

		if s3.Bucket != truth.bucket {
			t.Fatalf("Bucket was not parsed: %q %q %q", truth.url, truth.bucket, s3.Bucket)
		}

		if s3.Path != truth.path {
			t.Fatalf("Path was not parsed: %q %q %q", truth.url, truth.path, s3.Path)
		}

		if s3.Region != truth.region {
			t.Fatalf("Region was not parsed: %q %q %q", truth.url, truth.region, s3.Region)
		}

		if s3.AccessKey != truth.access_key {
			t.Fatalf("AccessKey was not parsed: %q %q %q", truth.url, truth.access_key, s3.AccessKey)
		}

		if s3.SecretKey != truth.secret_key {
			t.Fatalf("SecretKey was not parsed: %q %q %q", truth.url, truth.secret_key, s3.SecretKey)
		}
	}
}
