package helpers

import (
	"strings"

	"bitbucket.org/no-name-game/no-name/app/commands/godeffect/helpers/starnamer"
	"bitbucket.org/no-name-game/no-name/services"
	"github.com/mb-14/gomarkov"
)

// GenerateStarName - generate procedural name
func GenerateStarName() string {
	chain, err := starnamer.StarNameModel()
	if err != nil {
		services.ErrorHandler("Error load star name models", err)
	}

	order := chain.Order
	tokens := make([]string, 0)
	for i := 0; i < order; i++ {
		tokens = append(tokens, gomarkov.StartToken)
	}
	for tokens[len(tokens)-1] != gomarkov.EndToken {
		next, _ := chain.Generate(tokens[(len(tokens) - order):])
		tokens = append(tokens, next)
	}

	return strings.Join(tokens[order:len(tokens)-1], "")
}
