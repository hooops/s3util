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

## `s3://` url syntax

Several portions of the URL syntax are used for deriving the S3 information
used to handle the request. Several features are extensions of the `s3://`
syntax used in `s3cmd`.

* the `s3://` scheme must be provided.
* the username and password correspond to access and secret key (optional)
* The host is the name of the bucket, except for when...
* If a `.` is included in the host, the last portion of it will be used as the
  region (optional). If you wish to use bucket names with a dot in them, you
  will need to suffix with the region always.
* the path corresponds to the path in the bucket.

## Author

Erik Hollensbe <github@hollensbe.org>
