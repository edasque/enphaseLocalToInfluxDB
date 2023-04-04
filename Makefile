appname=enphaselocal2influx

all: run linux

linux: main.go
		GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build  -o $(appname).linux.amd64 main.go

macos: main.go
		GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build  -o $(appname).macos.amd64 main.go
		GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build  -o $(appname).macos.arm64 main.go
		lipo -create -output $(appname).macos $(appname).macos.amd64 $(appname).macos.arm64
		rm $(appname).macos.amd64 $(appname).macos.arm64
run: main.go
		go run main.go

clean:
		rm -f $(OUT)

docker-build: linux
		docker build . -t $(appname)

docker-run: docker-build
		docker run --rm -d --name $(appname) $(appname)