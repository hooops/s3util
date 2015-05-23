# S3Util: Parallel tools for interacting with S3

s3util provides a friendly CLI to parallel transfers.

Using a parallel transfer to S3 will maximize bandwidth usage. It can usually
hit close to wire speed on the AWS internal network. This is superior to
`s3cmd` for these scenarios. (`s3cmd` is great for everything else.)

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

### Examples

* `s3util put $PWD s3://my-bucket` will transfer `$PWD` to the my-bucket bucket.
* `s3util put $PWD s3://my-bucket.us-west-1/tmp` will transfer `$PWD` to the
  `my-bucket` bucket, folder `/tmp`, in region `us-west-1`.
* `s3util get s3://access:secret@my-bucket tmp` will get with the provided
  credentials (overriding the environment) and transfer from the root of the
  bucket to `$PWD/tmp`.

For other options such as concurrency, please see the `--help` option.

## Author

Erik Hollensbe <github@hollensbe.org>
