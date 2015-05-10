package get

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/codegangsta/cli"

	"github.com/erikh/s3util/common"
	"github.com/erikh/s3util/request"
	"github.com/erikh/s3util/s3url"
)

type GetConfig struct {
	Client     request.Client
	Pathchan   chan *string
	Donechan   chan struct{}
	BucketName string
	LocalPath  string
}

type Get struct {
	client   request.Client
	pathchan chan *string
	donechan chan struct{}
}

func NewGet() *Get {
	return &Get{
		client:   request.Client{},
		pathchan: make(chan *string),
		donechan: make(chan struct{}),
	}
}

func (g *Get) GetCommand(ctx *cli.Context) {
	if len(ctx.Args()) != 2 {
		common.ErrExit("Incorrect arguments. Try `%s --help`.", os.Args[0])
	}

	g.client = request.NewClient(
		ctx.String("access-key"),
		ctx.String("secret-key"),
		ctx.String("host"),
		ctx.String("region"),
	)

	if g.client.AWS.AccessKeyID == "" || g.client.AWS.SecretAccessKey == "" {
		fmt.Println("Invalid keys. Set AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY.")
		cli.ShowAppHelp(ctx)
		os.Exit(1)
	}

	s3url, err := s3url.ParseS3URL(ctx.Args()[0])
	if err != nil {
		fmt.Println(err)
		cli.ShowAppHelp(ctx)
		os.Exit(1)
	}

	bucketName := s3url.Bucket
	bucketPath := s3url.Path[1:]
	localPath := ctx.Args()[1]

	if bucketName == "" || localPath == "" {
		fmt.Println("Some flags are empty")
		flag.Usage()
		os.Exit(1)
	}

	if bucketPath == "/" || bucketPath == "." {
		bucketPath = ""
	}

	concurrency := ctx.Int("concurrency")
	for i := 0; i < concurrency; i++ {
		go g.fetch(bucketName, localPath)
	}

	marker := ""

	for {
		bucket := request.Bucket{}

		req, err := http.NewRequest("GET", fmt.Sprintf("https://%s.%s?marker=%s&prefix=%s", bucketName, g.client.Host, url.QueryEscape(marker), url.QueryEscape(bucketPath)), nil)
		if err != nil {
			common.ErrExit("Could not complete request: %v", err)
		}

		resp, err := g.client.Do(req)
		if err != nil {
			common.ErrExit("Could not complete request: %v", err)
		}

		if resp.StatusCode != 200 {
			fmt.Println("Ensure your region settings are correct.")
			common.ErrExit("Could not read bucket: fatal error.")
		}

		content, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			common.ErrExit("Failure during download: %v", err)
		}

		if err := xml.Unmarshal(content, &bucket); err != nil {
			common.ErrExit("Failure during XML parse: %v", err)
		}

		if len(bucket.Contents) == 0 {
			break
		}

		for _, item := range bucket.Contents {
			str := item.Key
			g.pathchan <- &str
		}

		marker = bucket.Contents[len(bucket.Contents)-1].Key
	}

	for i := 0; i < concurrency; i++ {
		g.pathchan <- nil
	}

	for i := 0; i < concurrency; i++ {
		<-g.donechan
	}
}

func (gc *GetConfig) Get() (bool, bool) {
	target := <-gc.Pathchan
	if target == nil {
		gc.Donechan <- struct{}{}
		return false, true
	}

	fullPath := filepath.Join(gc.LocalPath, *target)

	if strings.HasSuffix(*target, "/") {
		os.MkdirAll(fullPath, 0755)
		return true, false
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s.%s/%s", gc.BucketName, gc.Client.Host, *target), nil)
	if err != nil {
		gc.Pathchan <- target
		return false, false
	}

	resp, err := gc.Client.Do(req)
	if err != nil {
		gc.Pathchan <- target
		return false, false
	}

	if resp.StatusCode != 200 {
		fmt.Printf("Received status %d downloading; cannot continue.\n", resp.StatusCode)
		os.Exit(1)
	}

	// if there's any error requesting the url, retry.  silently do so if we
	// get an EOF error during the request. This usually means we're doing too
	// many requests at once.  be loud if we get some other error.

	// FIXME There's a race here between deleted files happening after the list
	// is performed. It lives in the channel for a while on large lists, so we
	// might need to stat it or react to the errors better or something. The
	// fail state here is to retry infinitely.
	if _, ok := err.(*url.Error); ok {
		gc.Pathchan <- target
		return false, false
	} else if err != nil {
		fmt.Printf("Error: %v - retrying\n", err)
		gc.Pathchan <- target
		return false, false
	}

	if err := os.MkdirAll(filepath.Dir(fullPath), 0700); err != nil {
		common.ErrExit("Could not create directory for download path: %v", err)
	}

	f, err := os.Create(fullPath)
	if err != nil {
		common.ErrExit("Could not create file: %v", err)
	}

	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		gc.Pathchan <- target
		return false, false
	}

	fmt.Println(*target, "->", fullPath)

	return true, false
}

func (g *Get) fetch(bucketName, localPath string) {
	for {
		gc := GetConfig{
			Client:     g.client,
			Pathchan:   g.pathchan,
			Donechan:   g.donechan,
			BucketName: bucketName,
			LocalPath:  localPath,
		}

		if ok, doBreak := gc.Get(); !ok && doBreak {
			break
		}

		time.Sleep(100 * time.Millisecond)
	}
}
