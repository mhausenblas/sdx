sdx_version := 0.1

.PHONY: build clean

###############################################################################
# commands related to Go testing and building binaries
build :
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.releaseVersion=$(sdx_version)" -o ./out/kube-sdx_macos *.go
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.releaseVersion=$(sdx_version)" -o ./out/kube-sdx_linux *.go
	GOOS=windows GOARCH=amd64 go build -ldflags "-X main.releaseVersion=$(sdx_version)" -o ./out/kube-sdx-_windows *.go

clean :
	@rm out/kube-sdx_*
