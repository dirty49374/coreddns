compile:
	go build -o coreddns main.go

run: compile
	./coreddns server ns1 -d lan --server ns1=10.0.1.31 --lease 10

publish:
	docker build -f Dockerfile . -t dirty49374/coreddns:$(version)
	docker build -f Dockerfile . -t dirty49374/coreddns:latest

	docker push dirty49374/coreddns:$(version)
	docker push dirty49374/coreddns:latest

compile-all:
	CGO_ENABLED=0 GOOS=windows go build -ldflags="-s -w" -a -installsuffix cgo -o ./build/windows/coreddns.exe .
	CGO_ENABLED=0 GOOS=darwin go build -ldflags="-s -w" -a -installsuffix cgo -o ./build/darwin/coreddns .
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -a -installsuffix cgo -o ./build/linux/coreddns .
