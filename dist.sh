#!/bin/sh

VERSION=$(cat VERSION)
mypwd=${PWD}

for os in linux darwin freebsd windows
do
	mkdir -p /tmp/s3util-${os}-${VERSION} && \
	cd /tmp/s3util-${os}-${VERSION} && \
	GOOS=$os GOARCH=amd64 go build github.com/erikh/s3util && \
	cd .. && \
	tar cvzf s3util-${os}-${VERSION}.tar.gz s3util-${os}-${VERSION} && \
  cp s3util-${os}-${VERSION}.tar.gz ${mypwd}	
done
