package env

import "os"

var ACCESS_KEY, SECRET_KEY, REGION string

func init() {
	ACCESS_KEY = os.Getenv("AWS_ACCESS_KEY_ID")
	if ACCESS_KEY == "" {
		ACCESS_KEY = os.Getenv("AWS_ACCESS_KEY")
	}

	SECRET_KEY = os.Getenv("AWS_SECRET_ACCESS_KEY")
	if SECRET_KEY == "" {
		SECRET_KEY = os.Getenv("AWS_SECRET_KEY")
	}

	REGION = os.Getenv("AWS_REGION")
}
