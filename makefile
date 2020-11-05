build:
	go build -o bin/quote -ldflags="-s -w" main.go

compress:
	upx --brute bin/quote

run:
	go run main.go