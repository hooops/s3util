package common

import (
	"fmt"
	"os"
	"time"
)

const BACKOFF = 100 * time.Millisecond

func ErrExit(msg string, args ...interface{}) {
	ErrWarn(msg, args...)
	os.Exit(1)
}

func ErrWarn(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
}

func TemplateHost(bucketName, host, target string) string {
	return fmt.Sprintf("https://%s.%s/%s", bucketName, host, target)
}
