all:
	cd src; go install -v ./...

rebuild:
	cd src; go install -a -v ./...

install: all
	@mkdir -p ~/.qbpl/formats
	cp formats/*.bpl ~/.qbpl/formats/

test:
	cd src; go test ./...

testv:
	cd src; GOOS=linux GOARCH=amd64 go test -v ./...

clean:
	cd ncgo; GOOS=linux GOARCH=amd64 go clean -i ./...

fmt:
	gofmt -w=true src/
