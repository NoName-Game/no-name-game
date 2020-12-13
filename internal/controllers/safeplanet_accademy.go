package controllers

import (
	"fmt"
	"strings"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// SafePlanetAccademyController
// ====================================
type SafePlanetAccademyController struct {
	Payload struct {
		AbilityCategoryID uint32
	}
	Controller
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetAccademyController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Verifico se è impossibile inizializzare
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.accademy",
			Payload:    &c.Payload,
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &MenuController{},
				FromStage: 1,
			},
			PlanetType: []string{"safe"},
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
	c.Completing(&c.Payload)
}

// ====================================
// Validator
// ====================================
func (c *SafePlanetAccademyController) Validator() (hasErrors bool) {
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Verifio se è stato passata una categorie tra quelle indicate
	// ##################################################################################################
	case 1:
		var err error

		// Recupero categorie abilità
		var rGetAllAbilityCategory *pb.GetAllAbilityCategoryResponse
		if rGetAllAbilityCategory, err = config.App.Server.Connection.GetAllAbilityCategory(helpers.NewContext(1), &pb.GetAllAbilityCategoryRequest{}); err != nil {
			c.Logger.Panic(err)
		}

		for _, category := range rGetAllAbilityCategory.GetAbilityCategories() {
			if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("safeplanet.accademy.%s", category.GetSlug())) {
				c.Payload.AbilityCategoryID = category.GetID()
				return false
			}
		}

		return true
	// ##################################################################################################
	//
	// ##################################################################################################
	case 2:
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "learn") {
			return true
		}
	}

	return false
}

// ====================================
// Stage
// ====================================
func (c *SafePlanetAccademyController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// Invio messaggio con recap stats
	case 0:
		var infoAccademy string
		infoAccademy = fmt.Sprintf("%s\n\n%s",
			helpers.Trans(c.Player.Language.Slug, "safeplanet.accademy.info"),
			helpers.Trans(c.Player.Language.Slug, "safeplanet.accademy.which"),
		)

		// Recupero categorie abilità
		var rGetAllAbilityCategory *pb.GetAllAbilityCategoryResponse
		if rGetAllAbilityCategory, err = config.App.Server.Connection.GetAllAbilityCategory(helpers.NewContext(1), &pb.GetAllAbilityCategoryRequest{}); err != nil {
			c.Logger.Panic(err)
		}

		var categoryKeyboard []tgbotapi.KeyboardButton
		for _, category := range rGetAllAbilityCategory.GetAbilityCategories() {
			categoryKeyboard = append(categoryKeyboard, tgbotapi.NewKeyboardButton(
				helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("safeplanet.accademy.%s", category.GetSlug())),
			))
		}

		msg := helpers.NewMessage(c.Player.ChatID, infoAccademy)
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			categoryKeyboard,
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.more")),
			),
		)

		msg.ParseMode = tgbotapi.ModeMarkdown
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		c.CurrentState.Stage = 1
	case 1:
		// Recupero abilità per la categoria scelta
		var rGetAbilityForPlayerByCategory *pb.GetAbilityForPlayerByCategoryResponse
		if rGetAbilityForPlayerByCategory, err = config.App.Server.Connection.GetAbilityForPlayerByCategory(helpers.NewContext(1), &pb.GetAbilityForPlayerByCategoryRequest{
			PlayerID:          c.Player.ID,
			AbilityCategoryID: c.Payload.AbilityCategoryID,
		}); err != nil {
			// Potrebbero esserci stati degli errori come per esempio la mancanza di amuleti
			errorMsg := helpers.NewMessage(c.Update.Message.Chat.ID,
				helpers.Trans(c.Player.Language.Slug, "safeplanet.accademy.no_more_abilities"),
			)
			if _, err = helpers.SendMessage(errorMsg); err != nil {
				c.Logger.Panic(err)
			}

			c.CurrentState.Completed = true
			c.Configurations.ControllerBack.To = &SafePlanetAccademyController{}
			return
		}

		// Costruisco messaggio con spiegazione abilità
		var abilityRecap string
		abilityRecap = helpers.Trans(c.Player.Language.Slug, "safeplanet.accademy.ability_details",
			// Nome
			rGetAbilityForPlayerByCategory.GetAbility().GetName(),
			// Livello
			rGetAbilityForPlayerByCategory.GetAbility().GetLevel(),
			// Descrizione
			helpers.Trans(c.Player.Language.Slug, fmt.Sprintf(
				"safeplanet.accademy.ability.%s.details", rGetAbilityForPlayerByCategory.GetAbility().GetSlug()),
			),
		)

		// Recupero quanti amuleti possiede il player
		var rGetPlayerAmulets *pb.GetPlayerAmuletsResponse
		if rGetPlayerAmulets, err = config.App.Server.Connection.GetPlayerAmulets(helpers.NewContext(1), &pb.GetPlayerAmuletsRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Costo
		abilityPrice := helpers.Trans(c.Player.Language.Slug, "safeplanet.accademy.ability_confirm",
			rGetPlayerAmulets.GetPlayerInventory().GetQuantity(),
			rGetAbilityForPlayerByCategory.GetAbility().GetAmulets(),
		)

		// Messaggio conferma finale
		msg := helpers.NewMessage(c.Player.ChatID, fmt.Sprintf("%s\n\n%s", abilityRecap, abilityPrice))
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "learn")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
			),
		)

		msg.ParseMode = tgbotapi.ModeMarkdown
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 2
	case 2:
		// Registro abilità al player
		_, err = config.App.Server.Connection.LearnAbility(helpers.NewContext(1), &pb.LearnAbilityRequest{
			PlayerID:          c.Player.ID,
			AbilityCategoryID: c.Payload.AbilityCategoryID,
		})

		if err != nil && strings.Contains(err.Error(), "not enough quantity") {
			// Potrebbero esserci stati degli errori come per esempio la mancanza di amuleti
			errorMsg := helpers.NewMessage(c.Update.Message.Chat.ID,
				helpers.Trans(c.Player.Language.Slug, "safeplanet.accademy.not_enough_amulets"),
			)
			if _, err = helpers.SendMessage(errorMsg); err != nil {
				c.Logger.Panic(err)
			}

			c.CurrentState.Completed = true
			c.Configurations.ControllerBack.To = &SafePlanetAccademyController{}
			return
		} else if err != nil {
			c.Logger.Panic(err)
		}

		// Mando messaggio di completamento
		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug,
			"safeplanet.accademy.done",
		))
		msg.ParseMode = tgbotapi.ModeMarkdown
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Completo lo stato
		c.CurrentState.Completed = true
		c.Configurations.ControllerBack.To = &SafePlanetAccademyController{}
	}

	return
}
