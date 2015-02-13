frontdesk: *.go
	go build .

run: frontdesk
	. .env && ./frontdesk

test: *.go
	go test .

install_deps:
	go get github.com/fluffle/goirc
	go get github.com/kelseyhightower/envconfig
	go get github.com/boltdb/bolt/
	go get github.com/gorilla/feeds
