client:
	@go run nn-telegram/cmd/client

notifier:
	@go run nn-telegram/cmd/notifier

lint:
	@golangci-lint run