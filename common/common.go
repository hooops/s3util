package common

import (
	"fmt"
	"os"
)

func ErrExit(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

func TemplateHost(bucketName, host, target string) string {
	return fmt.Sprintf("https://%s.%s/%s", bucketName, host, target)
}
