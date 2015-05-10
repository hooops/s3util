package get

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/codegangsta/cli"

	"github.com/erikh/s3util/bucket"
	"github.com/erikh/s3util/common"
	"github.com/erikh/s3util/request"
	"github.com/erikh/s3util/s3url"
)

const BACKOFF = 100 * time.Millisecond

type Get struct {
	bucketClient *bucket.BucketClient
	pathchan     chan *string
	donechan     chan struct{}
	localPath    string
	concurrency  int
}

func NewGet() *Get {
	return &Get{
		bucketClient: nil,
		pathchan:     make(chan *string),
		donechan:     make(chan struct{}),
	}
}

func (g *Get) GetCommand(ctx *cli.Context) {
	if err := g.handleArgs(ctx); err != nil {
		cli.ShowAppHelp(ctx)
		common.ErrExit(err.Error())
	}

	for i := 0; i < g.concurrency; i++ {
		go g.fetch()
	}

	if err := g.bucketClient.Find(g.push); err != nil {
		common.ErrExit(err.Error())
	}

	for i := 0; i < g.concurrency; i++ {
		g.pathchan <- nil
	}

	for i := 0; i < g.concurrency; i++ {
		<-g.donechan
	}
}

func (g *Get) handleArgs(ctx *cli.Context) error {
	if len(ctx.Args()) != 2 {
		return fmt.Errorf("Incorrect arguments. Try `%s --help`.", os.Args[0])
	}

	client := request.NewClient(
		ctx.String("access-key"),
		ctx.String("secret-key"),
		ctx.String("host"),
		ctx.String("region"),
	)

	if client.AWS.AccessKeyID == "" || client.AWS.SecretAccessKey == "" {
		return fmt.Errorf("Invalid keys. Set AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY.")
		cli.ShowAppHelp(ctx)
		os.Exit(1)
	}

	s3url, err := s3url.ParseS3URL(ctx.Args()[0])
	if err != nil {
		return err
	}

	localPath := ctx.Args()[1]

	if localPath == "" {
		return fmt.Errorf("Local path is missing. Please supply all arguments.")
	}

	g.localPath = localPath
	g.concurrency = ctx.Int("concurrency")

	g.bucketClient = bucket.NewBucketClient(s3url, client)

	return nil
}

func (g *Get) push(s *string) error {
	g.pathchan <- s
	return nil
}

func (g *Get) pushBack(s *string) {
	g.push(s)
	time.Sleep(BACKOFF)
}

func (g *Get) fetch() {
	for {
		target := <-g.pathchan
		if target == nil {
			g.donechan <- struct{}{}
			break
		}

		fullPath := filepath.Join(g.localPath, *target)

		if strings.HasSuffix(*target, "/") {
			os.MkdirAll(fullPath, 0755)
			continue
		}

		req, err := http.NewRequest("GET", common.TemplateHost(g.bucketClient.S3URL.Bucket, g.bucketClient.Client.Host, *target), nil)
		if err != nil {
			g.pushBack(target)
			continue
		}

		resp, err := g.bucketClient.Client.Do(req)
		if err != nil {
			g.pushBack(target)
			continue
		}

		if resp.StatusCode != 200 {
			common.ErrExit("Received status %d downloading; cannot continue.\n", resp.StatusCode)
		}

		// if there's any error requesting the url, retry.  silently do so if we
		// get an EOF error during the request. This usually means we're doing too
		// many requests at once.  be loud if we get some other error.

		// FIXME There's a race here between deleted files happening after the list
		// is performed. It lives in the channel for a while on large lists, so we
		// might need to stat it or react to the errors better or something. The
		// fail state here is to retry infinitely.
		if _, ok := err.(*url.Error); ok {
			g.pushBack(target)
			continue
		} else if err != nil {
			fmt.Printf("Error: %v - retrying\n", err)
			g.pushBack(target)
			continue
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
			g.pushBack(target)
			continue
		}

		fmt.Println(*target, "->", fullPath)
	}
}
