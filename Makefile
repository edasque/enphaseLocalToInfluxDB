all: run linux

linux: main.go
		GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build  -o enphaseLocalToInflux.linux.amd64 main.go

macos: main.go
		GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build  -o enphaseLocalToInflux.macos.amd64 main.go
		GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build  -o enphaseLocalToInflux.macos.arm64 main.go
		lipo -create -output enphaseLocalToInflux.macos enphaseLocalToInflux.macos.amd64 enphaseLocalToInflux.macos.arm64
run: main.go
		go run main.go

clean:
		rm -f $(OUT)

docker-build: linux
		docker build . -t enphaseLocalToInflux
		# docker-compose up --build -d postEngineeringJobsToSlack

docker-run: docker-build
		docker run --rm -d --name enphaseLocalToInflux enphaseLocalToInflux