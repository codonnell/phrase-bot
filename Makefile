run:
	@go run cmd/main.go

watch:
	@air --build.cmd "go build cmd/main.go" --build.bin "./main"
