.PHONY: dev prod test-nodocker

run:
	modd
test:
	go test -count=1 ./... 
test-nodocker:
	go test -count=1 -v -tags=nodocker ./...
latest:
	docker build -t docker.pkg.github.com/whitewater-guide/gorge/gorge:latest .
	docker push docker.pkg.github.com/whitewater-guide/gorge/gorge:latest
lint:
	golangci-lint run