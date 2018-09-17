linux:
	rm -rf ./dist
	GOOS=linux GOARCH=amd64 go build -o dist/ip-service-amd64 
	cp china_city.json dist/ && cp qqwry.dat dist/
dev:
	go run main.go routes.go QQWryReader.go 