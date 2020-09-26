package controllers

import (
	"time"

	"bitbucket.org/no-name-game/nn-telegram/config"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// ShipRestsController
// ====================================
type ShipRestsController struct {
	Controller
}

// ====================================
// Handle
// ====================================
func (c *ShipRestsController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Verifico se è impossibile inizializzare
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.ship.rests",
		},
		Configurations: ControllerConfigurations{
			ControllerBlocked: []string{"hunting", "mission"},
			ControllerBack: ControllerBack{
				To:        &ShipController{},
				FromStage: 1,
			},
		},
	}) {
		return
	}

	// Validate
	var hasError bool
	if hasError = c.Validator(); hasError {
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
func (c *ShipRestsController) Validator() (hasErrors bool) {
	switch c.CurrentState.Stage {
	case 0:
		// Verifico se il player necessita davvero di dormire
		restsInfo, err := config.App.Server.Connection.GetRestsInfo(helpers.NewContext(1), &pb.GetRestsInfoRequest{
			PlayerID: c.Player.GetID(),
		})
		if err != nil {
			c.Logger.Panic(err)
		}

		if !restsInfo.NeedRests {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "ship.rests.dont_need")
			return true
		}

		return false
	case 1:
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "ship.rests.start") {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")

			return true
		}

		return false
	case 2:
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "ship.rests.wakeup") {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "ship.rests.validator.need_to_wakeup")
			return true
		}

		// // Si vuole svegliare, ma non è passato ancora un minuto
		// var endDate = time.Now()
		//
		// diffDate := endDate.Sub(c.Payload.StartDateTime)
		// diffMinutes := math.RoundToEven(diffDate.Minutes())
		// if diffMinutes <= 1 {
		// 	c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "ship.rests.need_to_rest")
		// 	c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(
		// 		tgbotapi.NewKeyboardButtonRow(
		// 			tgbotapi.NewKeyboardButton(
		// 				helpers.Trans(c.Player.Language.Slug, "route.breaker.back"),
		// 			),
		// 		),
		// 	)
		// 	return true
		// }

		return false
	}

	return true
}

// ====================================
// Stage
// ====================================
func (c *ShipRestsController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// In questo riporto al player le tempistiche necesarie al riposo
	case 0:
		// Recupero informazioni per il recupero totale delle energie
		var restsInfo *pb.GetRestsInfoResponse
		if restsInfo, err = config.App.Server.Connection.GetRestsInfo(helpers.NewContext(1), &pb.GetRestsInfoRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Costruisco info per riposo
		var restsRecap string
		restsRecap = helpers.Trans(c.Player.Language.Slug, "ship.rests")
		if restsInfo.NeedRests {
			restsRecap += helpers.Trans(c.Player.Language.Slug, "ship.rests.time", restsInfo.GetRestsTime())
		}
		restsRecap += helpers.Trans(c.Player.Language.Slug, "ship.rests.info")

		// Aggiongo bottone start riposo
		var keyboardRow [][]tgbotapi.KeyboardButton
		if restsInfo.NeedRests {
			newKeyboardRow := tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "ship.rests.start"),
				),
			)
			keyboardRow = append(keyboardRow, newKeyboardRow)
		}

		// Aggiungo abbandona solo se il player non è morto e quindi obbligato a dormire
		if !c.Data.PlayerStats.GetDead() {
			keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
			))
		}

		// Invio messaggio
		msg := helpers.NewMessage(c.Player.ChatID, restsRecap)
		msg.ParseMode = "markdown"
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRow,
		}
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 1

	// In questo stage avvio effettivamente il riposo
	case 1:
		// Recupero informazioni per il recupero totale delle energie
		var rStartPlayerRest *pb.StartPlayerRestResponse
		if rStartPlayerRest, err = config.App.Server.Connection.StartPlayerRest(helpers.NewContext(1), &pb.StartPlayerRestRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Recupero orario fine riposo
		var finishAt time.Time
		if finishAt, err = helpers.GetEndTime(rStartPlayerRest.GetRestEndTime(), c.Player); err != nil {
			c.Logger.Panic(err)
		}

		// Invio messaggio
		msg := helpers.NewMessage(c.Update.Message.Chat.ID,
			helpers.Trans(c.Player.Language.Slug, "ship.rests.sleep", finishAt.Format("15:04:05")),
		)

		msg.ParseMode = "markdown"
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "ship.rests.wakeup")),
			),
		)

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 2
	case 2:
		// Fine riposo
		var rEndPlayerRest *pb.EndPlayerRestResponse
		if rEndPlayerRest, err = config.App.Server.Connection.EndPlayerRest(helpers.NewContext(1), &pb.EndPlayerRestRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Invio messaggio
		msg := helpers.NewMessage(c.Update.Message.Chat.ID,
			helpers.Trans(c.Player.Language.Slug, "ship.rests.finish", rEndPlayerRest.GetLifeRecovered()),
		)
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "route.breaker.more"),
				),
			),
		)

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Completo lo stato
		c.CurrentState.Completed = true
	}

	return
}
