client:
	@go run bitbucket.org/no-name-game/nn-telegram/cmd/client

notifier:
	@go run bitbucket.org/no-name-game/nn-telegram/cmd/notifier

lint:
	@golangci-lint run