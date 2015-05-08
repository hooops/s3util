# S3Util: Parallel tools for interacting with S3

s3util provides a friendly CLI to parallel transfers. Using a parallel
implementation of S3 uploading or downloading is the best way to get exemplary
throughput on small files.

s3util accepts s3cmd-style `s3://` urls for it's bucket references.

## Installation

```
$ go get github.com/erikh/s3util
```

## Usage

* `s3util help` will yield a help screen
* `s3util get --help` will show you how to use the get tool
* `s3util put --help` will show you how to use the put tool

**Note**: the default region is us-east-1 or the "classic" region. If you wish
to download files from another region, s3util will just download no files. This
is currently a bug in s3util, but if you specify `--region` you will find it
downloads your files properly.

## Author

Erik Hollensbe <github@hollensbe.org>
