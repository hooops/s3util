package put

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/codegangsta/cli"

	"github.com/erikh/s3util/common"
	"github.com/erikh/s3util/request"
	"github.com/erikh/s3util/s3url"
)

type putFile struct {
	request  *http.Request
	filename string
	url      string
}

type Put struct {
	s3url       s3url.S3URL
	client      request.Client
	requestChan chan *putFile
	doneChan    chan struct{}
	target      string
	concurrency int
}

func NewPut() *Put {
	return &Put{
		requestChan: make(chan *putFile),
		doneChan:    make(chan struct{}),
	}
}

func (p *Put) PutCommand(ctx *cli.Context) {
	if err := p.handleArgs(ctx); err != nil {
		cli.ShowAppHelp(ctx)
		common.ErrExit(err.Error())
	}

	for i := 0; i < p.concurrency; i++ {
		go p.runPut()
	}

	if err := filepath.Walk(p.target, p.handleFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	for i := 0; i < p.concurrency; i++ {
		p.requestChan <- nil
	}

	for i := 0; i < p.concurrency; i++ {
		<-p.doneChan
	}
}

func (p *Put) handleArgs(ctx *cli.Context) error {
	if len(ctx.Args()) != 2 {
		return fmt.Errorf("Invalid number of arguments.")
	}

	p.client = request.NewClient(
		ctx.String("access-key"),
		ctx.String("secret-key"),
		ctx.String("host"),
		ctx.String("region"),
	)

	p.target = ctx.Args()[0]

	s3url, err := s3url.ParseS3URL(ctx.Args()[1])
	if err != nil {
		return err
	}

	p.s3url = s3url

	if p.target == "" {
		return fmt.Errorf("Invalid target or bucket")
	}

	if p.client.AWS.AccessKeyID == "" || p.client.AWS.SecretAccessKey == "" {
		return fmt.Errorf("Invalid keys. Set AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY.")
	}

	p.concurrency = ctx.Int("concurrency")

	return nil
}

func (p *Put) runPut() {
	for {
		putfile := <-p.requestChan
		if putfile == nil {
			p.doneChan <- struct{}{}
			return
		}

		resp, err := p.client.Do(putfile.request)
		if err == nil {
			fmt.Printf("%s ~> %s\n", putfile.filename, putfile.url)
		} else {
			fmt.Printf("Error pushing %s: %s. retrying\n", putfile.filename, err)
			time.Sleep(100 * time.Millisecond)
			p.requestChan <- putfile
			// FIXME improve error
			if resp != nil && resp.Body != nil {
				io.Copy(os.Stdout, resp.Body)
			}
			continue
		}

		if resp.StatusCode != 200 {
			fmt.Printf("Received non-200 status code: %d. Cannot continue.\n", resp.StatusCode)
			fmt.Println("Ensure your region settings are correct.")
			os.Exit(1)
		}

		resp.Body.Close()
		putfile.request.Body.Close()

		time.Sleep(100 * time.Millisecond)
	}
}

func (p *Put) handleFile(path string, fi os.FileInfo, err error) error {
	if fi != nil && fi.IsDir() {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("Could not open %q for reading: %v", path, err)
	}

	sum := md5.New()

	buflen := 0

	for {
		buf := make([]byte, 4096)
		c, err := file.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}

		sum.Write(buf[:c])
		buflen += c

		if err == io.EOF {
			break
		}
	}

	remotePath := filepath.Clean(filepath.Join(p.s3url.Path, path))

	file.Seek(0, 0)

	url := common.TemplateHost(p.s3url.Bucket, p.client.Host, remotePath)
	req, err := http.NewRequest("PUT", url, file)
	if err != nil {
		return err
	}

	req.ContentLength = int64(buflen)
	req.Header.Add("Content-Type", "binary/octet-stream")
	req.Header.Add("Content-MD5", base64.StdEncoding.EncodeToString(sum.Sum(nil)))

	p.requestChan <- &putFile{
		request:  req,
		filename: path,
		url:      url,
	}

	return nil
}
