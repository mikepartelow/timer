timer: main.go
	@go build -o timer

install: timer
	@go install
