default:local

init:
	go mod tidy
	#go test ./...

local: init
	go install -trimpath -ldflags='-extldflags=-static -s -w' ./...

linux: init
	GOOS=linux GOARCH=amd64 go install -trimpath -ldflags='-extldflags=-static -s -w' ./...
	upx ~/go/bin/linux_amd64/wk
	upx ~/go/bin/linux_amd64/garnish
