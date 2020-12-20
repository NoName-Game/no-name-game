package controllers

import (
	"fmt"
	"strings"

	"bitbucket.org/no-name-game/nn-telegram/config"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// SafePlanetRelaxThreeCardController
// ====================================
type SafePlanetRelaxThreeCardController struct {
	Controller
	Payload struct {
		CardChoiced int32
	}
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetRelaxThreeCardController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Verifico se è impossibile inizializzare
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.relax.threecard",
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &SafePlanetRelaxController{},
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
	c.Completing(nil)
}

// ====================================
// Validator
// ====================================
func (c *SafePlanetRelaxThreeCardController) Validator() (hasErrors bool) {
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Verifico avvio missione
	// ##################################################################################################
	case 0:
		if helpers.Trans(c.Player.Language.Slug, "safeplanet.relax.threecards.play") == c.Update.Message.Text {
			c.CurrentState.Stage = 1
		}

	// ##################################################################################################
	// In questo stage andremo a verificare lo stato della missione
	// ##################################################################################################
	case 2:
		// Verifico subito se il player ha deciso di terminare la partita
		if helpers.Trans(c.Player.Language.Slug, "safeplanet.relax.threecards.play.give_up") == c.Update.Message.Text {
			c.CurrentState.Stage = 4
			return false
		}

		// Verifico se è stato passanto un numero
		verifyCards := map[int32]string{1: "1️⃣", 2: "2️⃣", 3: "3️⃣"}
		for i, card := range verifyCards {
			if card == c.Update.Message.Text {
				c.Payload.CardChoiced = i
				return false
			}
		}

		return true
	case 3:
		if helpers.Trans(c.Player.Language.Slug, "safeplanet.relax.threecards.play.continue") != c.Update.Message.Text {
			return true
		}
	}

	return false
}

// ====================================
// Stage
// ====================================
func (c *SafePlanetRelaxThreeCardController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Illustro gioco e chiedo se vuole giocare
	// ##################################################################################################
	case 0:
		var keyboardRows [][]tgbotapi.KeyboardButton
		keyboardRows = append(keyboardRows, []tgbotapi.KeyboardButton{
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "safeplanet.relax.threecards.play")),
		})

		// Aggiungo anche abbandona
		keyboardRows = append(keyboardRows, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(
				helpers.Trans(c.Player.Language.Slug, "route.breaker.more"),
			),
		))

		// Invio messaggi con il tipo di missioni come tastierino
		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.relax.threecards.intro"))
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			Keyboard:       keyboardRows,
			ResizeKeyboard: true,
		}
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

	// ##################################################################################################
	// Avvio partita e consegno prime carte
	// ##################################################################################################
	case 1:
		_, err = config.App.Server.Connection.ThreeCardGamePlay(helpers.NewContext(1), &pb.ThreeCardGamePlayRequest{
			PlayerID: c.Player.GetID(),
		})

		// Verifico se il player ha abbastanza soldi per giocare
		if err != nil && strings.Contains(err.Error(), "player dont have enough money") {
			// Potrebbero esserci stati degli errori come per esempio la mancanza di monete
			errorMsg := helpers.NewMessage(c.Update.Message.Chat.ID,
				helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.no_money"),
			)
			if _, err = helpers.SendMessage(errorMsg); err != nil {
				c.Logger.Panic(err)
			}

			c.CurrentState.Completed = true
			return
		} else if err != nil {
			c.Logger.Panic(err)
		}

		// Se non è esploso nulla allora posso dare le prime carte
		msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "safeplanet.relax.threecards.play.start"))
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("1️⃣"),
				tgbotapi.NewKeyboardButton("2️⃣"),
				tgbotapi.NewKeyboardButton("3️⃣"),
			),
		)
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Avanzo di stato
		c.CurrentState.Stage = 2

	// ##################################################################################################
	// Verifico carta scelta e premi
	// ##################################################################################################
	case 2:
		// Effettuo chiamata per verificare la giocata
		var rThreeCardGameCheckResponse *pb.ThreeCardGameCheckResponse
		if rThreeCardGameCheckResponse, err = config.App.Server.Connection.ThreeCardGameCheck(helpers.NewContext(1), &pb.ThreeCardGameCheckRequest{
			PlayerID:     c.Player.GetID(),
			PlayerChoice: c.Payload.CardChoiced,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Recupero possibile bottino
		var recapList = helpers.Trans(c.Player.Language.Slug, "safeplanet.relax.threecards.play.recap")
		for _, resource := range rThreeCardGameCheckResponse.GetResources() {
			recapList += fmt.Sprintf("- %s\n", resource.GetName())
		}

		// ***********************************
		// Lose
		// ***********************************
		if !rThreeCardGameCheckResponse.GetWin() {
			loseMessage := helpers.Trans(c.Player.Language.Slug, "safeplanet.relax.threecards.play.round_lose")

			// Invio messggio con ancora le 3 scelte di carte
			msg := helpers.NewMessage(c.Player.ChatID, fmt.Sprintf("%s \n%s", loseMessage, recapList))
			msg.ParseMode = tgbotapi.ModeMarkdown
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(
						helpers.Trans(c.Player.Language.Slug, "safeplanet.relax.threecards.play.continue"),
					),
				),
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(
						helpers.Trans(c.Player.Language.Slug, "route.breaker.clears"),
					),
				),
			)

			if _, err = helpers.SendMessage(msg); err != nil {
				c.Logger.Panic(err)
			}

			// Avanzo di stato
			c.CurrentState.Stage = 3
			return
		}

		// Recupero risorsa vinta in questo turno
		winMessage := helpers.Trans(c.Player.Language.Slug, "safeplanet.relax.threecards.play.round_win", rThreeCardGameCheckResponse.GetResource().GetName())

		// Invio messggio con ancora le 3 scelte di carte
		msg := helpers.NewMessage(c.Player.ChatID, fmt.Sprintf("%s \n%s", winMessage, recapList))
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("1️⃣"),
				tgbotapi.NewKeyboardButton("2️⃣"),
				tgbotapi.NewKeyboardButton("3️⃣"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "safeplanet.relax.threecards.play.give_up"),
				),
			),
		)

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

	// ##################################################################################################
	// Verifica se il player vuole usare i diamanti per continuare
	// ##################################################################################################
	case 3:
		// Se entro qui vuol dire che ha deciso di usare i diamanti
		_, err = config.App.Server.Connection.ThreeCardGameRecoverPlay(helpers.NewContext(1), &pb.ThreeCardGameRecoverPlayRequest{
			PlayerID: c.Player.GetID(),
		})

		// Verifico se il player ha abbastanza soldi per giocare
		if err != nil && strings.Contains(err.Error(), "player dont have enough diamond") {
			// Potrebbero esserci stati degli errori come per esempio la mancanza di monete
			errorMsg := helpers.NewMessage(c.Update.Message.Chat.ID,
				helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.no_money"),
			)
			if _, err = helpers.SendMessage(errorMsg); err != nil {
				c.Logger.Panic(err)
			}

			c.CurrentState.Completed = true
			return
		} else if err != nil {
			c.Logger.Panic(err)
		}

		// Invio messggio con ancora le 3 scelte di carte
		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.relax.threecards.play.recoverd"))
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("1️⃣"),
				tgbotapi.NewKeyboardButton("2️⃣"),
				tgbotapi.NewKeyboardButton("3️⃣"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "safeplanet.relax.threecards.play.give_up"),
				),
			),
		)

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Se il player decide di usare il diamante allora torna allo stato 2
		c.CurrentState.Stage = 2

	// ##################################################################################################
	// Il player ha deciso di terminare e portarsi a casa il bottino
	// ##################################################################################################
	case 4:
		// Richiamo ws per conclusione e mando messaggio se tutto ok
		if _, err = config.App.Server.Connection.ThreeCardGameEndGame(helpers.NewContext(1), &pb.ThreeCardGameEndGameRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.relax.threecards.play.end_game"))
		msg.ParseMode = tgbotapi.ModeMarkdown
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Completo lo stato
		c.CurrentState.Completed = true
	}

	return
}
