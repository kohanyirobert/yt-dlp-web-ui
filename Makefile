.PHONY : fe clean all

default:
	go run main.go

fe:
	cd frontend && pnpm install && pnpm build

dev:
	cd frontend && pnpm install && pnpm dev

all: fe
	CGO_ENABLED=0 go build -o yt-dlp-web-ui main.go

multiarch: fe
	mkdir -p build
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/yt-dlp-web-ui_linux-amd64 main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o build/yt-dlp-web-ui_linux-arm64 main.go
	CGO_ENABLED=0 GOOS=linux GOARM=6 GOARCH=arm go build -o build/yt-dlp-web-ui_linux-armv6 main.go
	CGO_ENABLED=0 GOOS=linux GOARM=7 GOARCH=arm go build -o build/yt-dlp-web-ui_linux-armv7 main.go

clean:
	rm -rf build
