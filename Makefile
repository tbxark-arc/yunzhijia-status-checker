.PHONY: buildLinuxX86
buildLinuxX86:
	GOOS=linux GOARCH=amd64 go build -o ./build/ ./...

.PHONY: buildImage
buildImage: buildLinuxX86
	docker buildx build --platform=linux/amd64 -t ghcr.io/tbxark-arc/yunzhijia-status-checker:latest .
	docker push ghcr.io/tbxark-arc/yunzhijia-status-checker:latest