package put

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
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
	wg          sync.WaitGroup
	target      string
	concurrency int
}

func NewPut() *Put {
	return &Put{
		wg:          sync.WaitGroup{},
		requestChan: make(chan *putFile),
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

	p.wg.Wait()
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
	p.wg.Add(1)

	for {
		putfile := <-p.requestChan
		if putfile == nil {
			p.wg.Done()
			return
		}

		resp, err := p.client.Do(putfile.request)
		if err != nil {
			common.ErrWarn("Error pushing %s: %s. retrying", putfile.filename, err)
			time.Sleep(common.BACKOFF)
			p.requestChan <- putfile
			continue
		}

		fmt.Printf("%s ~> %s\n", putfile.filename, putfile.url)

		if resp.StatusCode != 200 {
			common.ErrWarn("Received non-200 status code: %d. Cannot continue.", resp.StatusCode)
			common.ErrWarn("Ensure your region settings are correct.")
			os.Exit(1)
		}

		resp.Body.Close()
		putfile.request.Body.Close()

		time.Sleep(common.BACKOFF)
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

	md5sum, buflen, err := p.sumFile(file)
	if err != nil {
		return err
	}

	remotePath := filepath.Clean(filepath.Join(p.s3url.Path, path))

	file.Seek(0, 0)

	req, url, err := p.createPut(file, md5sum, buflen, remotePath)
	if err != nil {
		return err
	}

	p.requestChan <- &putFile{
		request:  req,
		filename: path,
		url:      url,
	}

	return nil
}

func (p *Put) createPut(file io.ReadCloser, md5sum string, md5len int, remotePath string) (*http.Request, string, error) {
	url := common.TemplateHost(p.s3url.Bucket, p.client.Host, remotePath)
	req, err := http.NewRequest("PUT", url, file)
	if err != nil {
		return nil, "", err
	}

	req.ContentLength = int64(md5len)
	req.Header.Add("Content-Type", "binary/octet-stream")
	req.Header.Add("Content-MD5", md5sum)

	return req, url, nil
}

func (p *Put) sumFile(file io.ReadCloser) (string, int, error) {
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
