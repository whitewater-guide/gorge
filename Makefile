download:
	go mod download

tools: download
	cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %

generate: tools
	go-bindata -o storage/migrations/sqlite/sqlite_migrations.go -pkg sqlite_migrations -prefix "storage/migrations/sqlite/" storage/migrations/sqlite/
	go-bindata -o storage/migrations/postgres/postgres_migrations.go -pkg postgres_migrations -prefix "storage/migrations/postgres/" storage/migrations/postgres/

build: GOOS=linux 
build: generate
	govvv build -o /go/bin/gorge-server -ldflags="-s -w -X main.BuildNumber=$(VERSION)" github.com/whitewater-guide/gorge/server
	govvv build -o /go/bin/gorge-cli -ldflags="-s -w -X main.BuildNumber=$(VERSION)" github.com/whitewater-guide/gorge/cli

test: generate
	go test -count=1 ./... 

test-nodocker: generate
	go test -count=1 -v -tags=nodocker ./...

lint: tools
	golangci-lint run

typescript: tools
	go run ./typescriptify

run: tools
	modd

######################################################
# ⭡ Command above run inside docker container       #
# ⭣ Command below run outside of docker container   #
######################################################
verify:
	docker build --target tester -t gorge_tester .
	# Extract index.d.ts from image
	docker run --rm --entrypoint "/bin/sh" \
           -v $(shell pwd):/extract gorge_tester -c "cp index.d.ts /extract"
prepare:
	docker build --build-arg VERSION=$(VERSION) -t docker.pkg.github.com/whitewater-guide/gorge/gorge:latest .
	docker tag docker.pkg.github.com/whitewater-guide/gorge/gorge:latest docker.pkg.github.com/whitewater-guide/gorge/gorge:$(VERSION)
publish:
	docker login docker.pkg.github.com -u $(GITHUB_USER) -p $(GITHUB_TOKEN)
	docker push docker.pkg.github.com/whitewater-guide/gorge/gorge:latest
	docker push docker.pkg.github.com/whitewater-guide/gorge/gorge:$(VERSION)