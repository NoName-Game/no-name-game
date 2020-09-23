module bitbucket.org/no-name-game/nn-telegram

go 1.14

replace bitbucket.org/no-name-game/nn-grpc => ../nn-grpc

require (
	bitbucket.org/no-name-game/nn-grpc v1.0.0
	github.com/certifi/gocertifi v0.0.0-20200922220541-2c3bb06c6054 // indirect
	github.com/evalphobia/logrus_sentry v0.8.2
	github.com/getsentry/raven-go v0.2.0 // indirect
	github.com/getsentry/sentry-go v0.5.1
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/go-telegram-bot-api/telegram-bot-api v4.6.4+incompatible
	github.com/golang/protobuf v1.4.2
	github.com/joho/godotenv v1.3.0
	github.com/labstack/gommon v0.3.0
	github.com/nicksnyder/go-i18n/v2 v2.0.2
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.6.0
	github.com/technoweenie/multipartstreamer v1.0.1 // indirect
	golang.org/x/text v0.3.3
	google.golang.org/grpc v1.31.1
	gopkg.in/yaml.v2 v2.2.4
)
