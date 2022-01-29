#!/bin/sh

cross_arm64() {
	for os in linux windows darwin freebsd openbsd netbsd; do
		echo $os
		CGO_ENABLED=0 GOMIPS=softfloat GOOS=$os GOARCH=arm64 go build -o ./goben/goben_${os}_arm64_softfloat ./goben
	done
}

cross_arm64
