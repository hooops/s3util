package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/codegangsta/cli"
	"github.com/erikh/s3util/common"
)

type get struct {
	client   *http.Client
	pathchan chan *string
	donechan chan struct{}
}

func newget() *get {
	return &get{
		client:   &http.Client{},
		pathchan: make(chan *string),
		donechan: make(chan struct{}),
	}
}

func (g *get) fetch(host, bucketName, localPath string) {
	for {
		gc := common.GetConfig{
			Client:     g.client,
			Pathchan:   g.pathchan,
			Donechan:   g.donechan,
			BucketName: bucketName,
			LocalPath:  localPath,
			Host:       host,
		}

		if ok, doBreak := gc.Get(); !ok && doBreak {
			break
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func (g *get) getCommand(ctx *cli.Context) {
	host, region := ctx.String("host"), ctx.String("region")

	if host != "" && region != "" {
		fmt.Println("You cannot set both host and region.")
		cli.ShowAppHelp(ctx)
		os.Exit(1)
	}

	if len(ctx.Args()) != 2 {
		common.ErrExit("Incorrect arguments. Try `%s --help`.", os.Args[0])
	}

	if common.ACCESS_KEY == "" || common.SECRET_KEY == "" {
		fmt.Println("Invalid keys. Set AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY.")
		cli.ShowAppHelp(ctx)
		os.Exit(1)
	}

	s3url, err := common.ParseS3URL(ctx.Args()[0])
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

	marker := ""
	concurrency := ctx.Int("concurrency")

	myhost := host
	if myhost == "" {
		if region != "" {
			myhost = fmt.Sprintf("s3-%s.amazonaws.com", region)
		} else {
			myhost = "s3.amazonaws.com"
		}
	}

	for i := 0; i < concurrency; i++ {
		go g.fetch(myhost, bucketName, localPath)
	}

	for {
		bucket := common.Bucket{}

		req, err := http.NewRequest("GET", fmt.Sprintf("https://%s.%s?marker=%s&prefix=%s", bucketName, myhost, url.QueryEscape(marker), url.QueryEscape(bucketPath)), nil)
		if err != nil {
			common.ErrExit("Could not complete request: %v", err)
		}

		resp, err := common.Request(g.client, req)
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
