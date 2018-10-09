sdx_version := 0.1

.PHONY: build clean

###############################################################################
# commands related to Go testing and building binaries
build :
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.releaseVersion=$(sdx_version)" -o ./out/sdx_macos *.go
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.releaseVersion=$(sdx_version)" -o ./out/sdx_linux *.go
	GOOS=windows GOARCH=amd64 go build -ldflags "-X main.releaseVersion=$(sdx_version)" -o ./out/sdx_windows *.go

clean :
	@rm out/sdx_*
