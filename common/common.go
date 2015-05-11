package common

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
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

func SumFile(file io.ReadCloser) (string, int, error) {
	sum := md5.New()

	buflen := 0

	for {
		buf := make([]byte, 4096)
		c, err := file.Read(buf)
		if err != nil && err != io.EOF {
			return "", 0, err
		}

		sum.Write(buf[:c])
		buflen += c

		if err == io.EOF {
			break
		}
	}

	md5sum := base64.StdEncoding.EncodeToString(sum.Sum(nil))

	return md5sum, buflen, nil
}
