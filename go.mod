module bitbucket.org/no-name-game/nn-telegram

go 1.14

replace bitbucket.org/no-name-game/nn-grpc => ../nn-grpc

require (
	bitbucket.org/no-name-game/nn-grpc v1.0.0
	github.com/certifi/gocertifi v0.0.0-20180118203423-deb3ae2ef261
	github.com/getsentry/sentry-go v0.5.1
	github.com/go-redis/redis v6.15.5+incompatible
	github.com/go-telegram-bot-api/telegram-bot-api v4.6.4+incompatible
	github.com/golang/protobuf v1.4.2
	github.com/joho/godotenv v1.3.0
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/nicksnyder/go-i18n v2.0.2+incompatible // indirect
	github.com/nicksnyder/go-i18n/v2 v2.0.2
	github.com/sirupsen/logrus v1.4.2
	github.com/technoweenie/multipartstreamer v1.0.1 // indirect
	golang.org/x/sys v0.0.0-20190927073244-c990c680b611 // indirect
	golang.org/x/text v0.3.2
	google.golang.org/grpc v1.30.0
	gopkg.in/yaml.v2 v2.2.4
)
