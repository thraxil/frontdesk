frontdesk: *.go
	go build .

run: frontdesk
	. .env && ./frontdesk

test: *.go
	go test .

coverage: frontdesk
	go test . -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

install_deps:
	go get github.com/fluffle/goirc
	go get github.com/kelseyhightower/envconfig
	go get github.com/boltdb/bolt/
	go get github.com/gorilla/feeds
	go get github.com/abbot/go-http-auth
	go get github.com/blevesearch/bleve
