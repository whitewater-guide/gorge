download:
	go mod download

tools: download
	cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %

generate: tools
	go-bindata -o storage/migrations/sqlite/sqlite_migrations.go -pkg sqlite_migrations -prefix "storage/migrations/sqlite/" storage/migrations/sqlite/
	go-bindata -o storage/migrations/postgres/postgres_migrations.go -pkg postgres_migrations -prefix "storage/migrations/postgres/" storage/migrations/postgres/

build: GOOS=linux 
build: generate
	go build -o /go/bin/gorge-server -ldflags="-s -w" github.com/whitewater-guide/gorge/server
	go build -o /go/bin/gorge-cli -ldflags="-s -w" github.com/whitewater-guide/gorge/cli

test: generate
	go test -count=1 ./... 

test-nodocker: generate
	go test -count=1 -v -tags=nodocker ./...

lint: tools
	golangci-lint run

run: tools
	modd

######################################################
# ⭡ Command above run inside docker container       #
# ⭣ Command below run outside of docker container   #
######################################################
verify:
	docker build --target tester .
prepare:
	docker build -t docker.pkg.github.com/whitewater-guide/gorge/gorge:latest .
	docker tag docker.pkg.github.com/whitewater-guide/gorge/gorge:latest docker.pkg.github.com/whitewater-guide/gorge/gorge:$(VERSION)
publish: latest
	docker push docker.pkg.github.com/whitewater-guide/gorge/gorge:latest
	docker push docker.pkg.github.com/whitewater-guide/gorge/gorge:$(VERSION)