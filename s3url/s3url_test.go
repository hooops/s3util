package s3url

import "testing"

type truth struct {
	url      string
	succeeds bool
	bucket   string
	path     string
	region   string
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
	}
}
