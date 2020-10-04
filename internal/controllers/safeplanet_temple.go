package controllers

import (
	"fmt"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// SafePlanetTempleController
// ====================================
type SafePlanetTempleController struct {
	Controller
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetTempleController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Verifico se è impossibile inizializzare
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.temple",
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &MenuController{},
				FromStage: 1,
			},
		},
	}) {
		return
	}

	// Validate
	if c.Validator() {
		c.Validate()
		return
	}

	// Ok! Run!
	c.Stage()

	// Completo progressione
	c.Completing(nil)
}

// ====================================
// Validator
// ====================================
func (c *SafePlanetTempleController) Validator() (hasErrors bool) {
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Verifio se è stato passata una categorie tra quelle indicate
	// ##################################################################################################
	case 1:
		return false
	// ##################################################################################################
	//
	// ##################################################################################################
	case 2:
		// TODO: Verificare importo
		return false
	}

	return false
}

// ====================================
// Stage
// ====================================
func (c *SafePlanetTempleController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// Invio messaggio con recap stats
	case 0:
		var infoTemple string
		infoTemple = helpers.Trans(c.Player.Language.Slug, "safeplanet.temple.info")

		// Recupero categorie abilità
		var rGetAllAbilityCategory *pb.GetAllAbilityCategoryResponse
		if rGetAllAbilityCategory, err = config.App.Server.Connection.GetAllAbilityCategory(helpers.NewContext(1), &pb.GetAllAbilityCategoryRequest{}); err != nil {
			c.Logger.Panic(err)
		}

		var categoryKeyboard []tgbotapi.KeyboardButton
		for _, category := range rGetAllAbilityCategory.GetAbilityCategories() {
			categoryKeyboard = append(categoryKeyboard, tgbotapi.NewKeyboardButton(
				helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("safeplanet.temple.%s", category.GetSlug())),
			))
		}

		msg := helpers.NewMessage(c.Player.ChatID, infoTemple)
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			categoryKeyboard,
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.more")),
			),
		)

		msg.ParseMode = "markdown"
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		c.CurrentState.Stage = 1
	case 1:
		// Recupero abilità per la categoria scelta
		var rGetAbilityForPlayerByCategory *pb.GetAbilityForPlayerByCategoryResponse
		if rGetAbilityForPlayerByCategory, err = config.App.Server.Connection.GetAbilityForPlayerByCategory(helpers.NewContext(1), &pb.GetAbilityForPlayerByCategoryRequest{
			PlayerID:          c.Player.ID,
			AbilityCategoryID: 1,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Costruisco messaggio con spiegazione abilità
		var abilityRecap string
		abilityRecap = helpers.Trans(c.Player.Language.Slug, "safeplanet.temple.ability_details", rGetAbilityForPlayerByCategory.GetAbility().GetName())

		msg := helpers.NewMessage(c.Player.ChatID, abilityRecap)
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
			),
		)

		msg.ParseMode = "markdown"
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		c.CurrentState.Stage = 1

		// Aggiorno stato
		c.CurrentState.Stage = 2
	case 2:

		// Completo lo stato
		c.CurrentState.Completed = true
		c.Configurations.ControllerBack.To = &SafePlanetTempleController{}
	}

	return
}
