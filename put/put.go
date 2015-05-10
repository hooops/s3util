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

	"github.com/erikh/s3util/env"
	"github.com/erikh/s3util/request"
	"github.com/erikh/s3util/s3url"
)

type putFile struct {
	request  *http.Request
	filename string
	url      string
}

type Put struct {
	client      *http.Client
	requestChan chan *putFile
	doneChan    chan struct{}
}

func NewPut() *Put {
	return &Put{
		client:      new(http.Client),
		requestChan: make(chan *putFile),
		doneChan:    make(chan struct{}),
	}
}

func (p *Put) PutCommand(ctx *cli.Context) {
	if len(ctx.Args()) != 2 {
		cli.ShowAppHelp(ctx)
		os.Exit(1)
	}

	target := ctx.Args()[0]
	s3url, err := s3url.ParseS3URL(ctx.Args()[1])
	if err != nil {
		fmt.Println(err)
		cli.ShowAppHelp(ctx)
		os.Exit(1)
	}

	if target == "" {
		fmt.Println("Invalid target or bucket")
		cli.ShowAppHelp(ctx)
		os.Exit(1)
	}

	env.Init(ctx.String("access-key"), ctx.String("secret-key"))

	if env.ACCESS_KEY == "" || env.SECRET_KEY == "" {
		fmt.Println("Invalid keys. Set AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY.")
		cli.ShowAppHelp(ctx)
		os.Exit(1)
	}

	concurrency := ctx.Int("concurrency")

	p.startPutters(concurrency)

	err = filepath.Walk(target, func(path string, fi os.FileInfo, err error) error {
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

		remotePath := filepath.Clean(filepath.Join(s3url.Path, path))

		file.Seek(0, 0)

		myhost := ctx.String("host")

		if myhost == "" {
			region := ctx.String("region")

			if region != "" {
				myhost = fmt.Sprintf("s3-%s.amazonaws.com", region)
			} else {
				myhost = "s3.amazonaws.com"
			}
		}

		url := fmt.Sprintf("https://%s.%s%s", s3url.Bucket, myhost, remotePath)
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
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	for i := 0; i < concurrency; i++ {
		p.requestChan <- nil
	}

	for i := 0; i < concurrency; i++ {
		<-p.doneChan
	}
}

func (p *Put) runPut() {
	for {
		putfile := <-p.requestChan
		if putfile == nil {
			p.doneChan <- struct{}{}
			return
		}

		resp, err := request.Request(p.client, putfile.request)
		if err == nil {
			fmt.Printf("%s ~> %s\n", putfile.filename, putfile.url)
		} else {
			fmt.Print("Error receiving %s: %s. retrying", putfile.filename, err)
			p.requestChan <- putfile
			// FIXME improve error
			if resp != nil && resp.Body != nil {
				io.Copy(os.Stdout, resp.Body)
			}
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

func (p *Put) startPutters(concurrency int) {
	for i := 0; i < concurrency; i++ {
		go p.runPut()
	}
}
