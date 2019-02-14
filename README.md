# NoName - Telegram Bot Game

This is THE project of NoName. NoName it's a telegram bot game.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

No-Name requires:

- [Docker](https://www.docker.com/) - Needed for run containers with postegres DB and Redis DB
- [Go](https://golang.org/) - Simple.. it's language of bot.
- [Glide](https://github.com/Masterminds/glide/blob/master/README.md) - Glide it's package manager used in this project, for install and manage dependencies.

### Installing

Create a copy of .env.example renamed to .env and set up the parameters.

- Set up the Telegram Bot Token "TELEGRAM_APIKEY"
- Set up the database info (To use docker container see the needed parameters in docker-compose.yaml)

Now update packages via Glide by typing:

```sh
$ glide up
```

once the dependencies are downloaded, we can run the docker for starting database containers, then run:

```sh
$ docker-compose up -d
```

The last step it's run application:

```sh
$ go run main.go
```

## Running/Make the tests

//

## Coding style

NoName use [gometalinter](https://github.com/alecthomas/gometalinter), this linter it's configured in pipelines and run by default at every pull request.
If you use VSCode like IDE i advice you to add this configuration in your settings.

```
  "go.lintTool": "gometalinter",
  "go.lintOnSave": "file",
  "go.lintFlags": [
    "--vendor",
    "--tests",
    "--cyclo-over=50",
    "--disable-all",
    "--enable=vet",
    "--enable=goimports",
    "--enable=vetshadow",
    "--enable=golint",
    "--enable=ineffassign",
    "--enable=goconst",
    "--enable=dupl",
    "--enable=gocyclo"
  ]
```

## Deployment

Add additional notes about how to deploy this on a live system

## Built With

- [Telegram-Bot-Api](https://github.com/go-telegram-bot-api/telegram-bot-api) - The telegram framework

## Contributing

NoName project use git-flow system.
To create a new feature just clone the repo on your local machine and initialize git flow and create a new feature (Look at this [link](https://danielkummer.github.io/git-flow-cheatsheet/) to see git flow commands).

## Versioning

We use [SemVer](http://semver.org/) for versioning.

## Authors / Contributors

- **Vito castellano** - _NoName Creator & Developer_
- **Lorenzo Avallone** - _Developer_

## License

This project is private! Any distribution without permission of the creator can be legally punished.
