package controllers

import (
	"fmt"
	"strconv"
	"strings"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// SafePlanetResearchDonationController
// ====================================
type SafePlanetResearchDonationController struct {
	Payload struct {
		ResourceID uint32
		Quantity   int32
	}
	Controller
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetResearchDonationController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Verifico se è impossibile inizializzare
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.coalition.research.donation",
			Payload:    &c.Payload,
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &SafePlanetResearchController{},
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
func (c *SafePlanetResearchDonationController) Validator() (hasErrors bool) {
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Verifico se la risorsa passata esiste
	// ##################################################################################################
	case 1:
		resourceName := strings.Split(c.Update.Message.Text, " (")[0]

		var err error
		var rGetResourceByName *pb.GetResourceByNameResponse
		if rGetResourceByName, err = config.App.Server.Connection.GetResourceByName(helpers.NewContext(1), &pb.GetResourceByNameRequest{
			Name: resourceName,
		}); err != nil {
			return true
		}

		c.Payload.ResourceID = rGetResourceByName.GetResource().GetID()

	// ##################################################################################################
	// Verifico quantità di item
	// ##################################################################################################
	case 2:
		quantity, _ := strconv.Atoi(c.Update.Message.Text)
		if quantity <= 0 {
			return true
		}

		c.Payload.Quantity = int32(quantity)

	// ##################################################################################################
	// Verifico la conferma dell'uso
	// ##################################################################################################
	case 3:
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "confirm") {
			return true
		}
	}

	return false
}

// ====================================
// Stage
// ====================================
func (c *SafePlanetResearchDonationController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	case 0:
		// Recupero informazioni ricerca
		var err error
		var rGetRecapActiveResearch *pb.GetRecapActiveResearchResponse
		if rGetRecapActiveResearch, err = config.App.Server.Connection.GetRecapActiveResearch(helpers.NewContext(1), &pb.GetRecapActiveResearchRequest{}); err != nil {
			c.Logger.Panic(err)
		}

		// Recupero risorse del player
		var rGetPlayerResources *pb.GetPlayerResourcesResponse
		if rGetPlayerResources, err = config.App.Server.Connection.GetPlayerResources(helpers.NewContext(1), &pb.GetPlayerResourcesRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Costruisco keyboard con le risorse
		var keyboardRow [][]tgbotapi.KeyboardButton
		for _, resource := range rGetPlayerResources.GetPlayerInventory() {
			// Filtro per la stessa raarità della ricerca in corso
			if resource.GetResource().GetRarityID() == rGetRecapActiveResearch.GetResearch().GetDonationRarityID() {
				newKeyboardRow := tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(
						fmt.Sprintf("%s (%v)", resource.GetResource().GetName(), resource.Quantity),
					),
				)
				keyboardRow = append(keyboardRow, newKeyboardRow)
			}
		}

		// Aggiungo torna indietro
		keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
		))

		// Invio messaggio
		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.research.donation.choose_reseource"))
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRow,
		}
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 1

	case 1:
		// Chiedo di indicare la quantità
		var keyboardRowQuantities [][]tgbotapi.KeyboardButton
		for i := 1; i <= 5; i++ {
			keyboardRow := tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(fmt.Sprintf("%d", i)),
			)
			keyboardRowQuantities = append(keyboardRowQuantities, keyboardRow)
		}

		// Aggiungo tasti back and clears
		keyboardRowQuantities = append(keyboardRowQuantities, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
		))

		// Invio messaggio
		msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.research.donation.choose_quantity"))
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowQuantities,
		}

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 2
	case 2:
		var rGetResourceByID *pb.GetResourceByIDResponse
		if rGetResourceByID, err = config.App.Server.Connection.GetResourceByID(helpers.NewContext(1), &pb.GetResourceByIDRequest{
			ID: c.Payload.ResourceID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Invio messaggio
		msg := helpers.NewMessage(c.Update.Message.Chat.ID,
			helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.research.donation.confirmation",
				c.Payload.Quantity,
				rGetResourceByID.GetResource().GetName(),
			),
		)
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "confirm")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
			),
		)

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 3
	case 3:
		// Concludo donazione
		_, err = config.App.Server.Connection.ResearchDonation(helpers.NewContext(1), &pb.ResearchDonationRequest{
			PlayerID:   c.Player.ID,
			ResourceID: c.Payload.ResourceID,
			Quantity:   c.Payload.Quantity,
		})

		if err != nil && strings.Contains(err.Error(), "player dont have enough resource quantity") {
			// Potrebbero esserci stati degli errori come per esempio la mancanza di materie prime
			errorMsg := helpers.NewMessage(c.Update.Message.Chat.ID,
				helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.research.donation.not_enough_resource"),
			)
			if _, err = helpers.SendMessage(errorMsg); err != nil {
				c.Logger.Panic(err)
			}

			c.CurrentState.Completed = true
			return
		}

		// Invio messaggio
		msg := helpers.NewMessage(c.Update.Message.Chat.ID,
			helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.research.donation.completed"),
		)

		msg.ParseMode = tgbotapi.ModeMarkdown
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Completo lo stato
		c.CurrentState.Completed = true
	}

	return
}
