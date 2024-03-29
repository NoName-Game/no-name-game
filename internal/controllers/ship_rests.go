package controllers

import (
	"time"

	"nn-telegram/config"

	"nn-grpc/build/pb"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"nn-telegram/internal/helpers"
)

// ====================================
// ShipRestsController
// ====================================
type ShipRestsController struct {
	Controller
}

func (c *ShipRestsController) Configuration(player *pb.Player, update tgbotapi.Update) Controller {
	return Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.ship.rests",
		},
		Configurations: ControllerConfigurations{
			ControllerBlocked: []string{"exploration"},
			ControllerBack: ControllerBack{
				To:        &ShipController{},
				FromStage: 1,
			},
			BreakerPerStage: map[int32][]string{
				0: {"route.breaker.menu"},
				1: {"route.breaker.menu"},
			},
			AllowedControllers: []string{
				"ship.rests.wakeup",
			},
		},
	}
}

// ====================================
// Handle
// ====================================
func (c *ShipRestsController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Verifico se è impossibile inizializzare
	if !c.InitController(c.Configuration(player, update)) {
		return
	}

	// Validate
	if c.Validator() {
		c.Validate()
		c.Completing(nil)
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
	// NNT-140 Se il player arriva dalla caccia (quindi callback query) mando direttamente allo stage 0
	if c.Update.CallbackQuery != nil {
		c.CurrentState.Stage = 0
		return false
	}

	// Se è stato passato il comando sveglia e il player sta effettivamente dormendo lo sveglio
	if helpers.Trans(c.Player.Language.Slug, "ship.rests.wakeup") == c.Update.Message.Text {
		var rRestCheck *pb.RestCheckResponse
		if rRestCheck, _ = config.App.Server.Connection.RestCheck(helpers.NewContext(1), &pb.RestCheckRequest{
			PlayerID: c.Player.GetID(),
		}); rRestCheck.GetInRest() {
			c.CurrentState.Stage = 1
		}
	}

	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Verifico se il player necessita davvero di dormire
	// ##################################################################################################
	case 0:
		var err error
		var rGetRestsInfo *pb.GetRestsInfoResponse
		if rGetRestsInfo, err = config.App.Server.Connection.GetRestsInfo(helpers.NewContext(1), &pb.GetRestsInfoRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		if !rGetRestsInfo.GetNeedRests() {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "ship.rests.dont_need")
			c.CurrentState.Completed = true
			return true
		}

	// ##################################################################################################
	// Verifico se il player vuole svegliarsi
	// ##################################################################################################
	case 1:
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "ship.rests.wakeup") {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "ship.rests.validator.need_to_wakeup")
			return true
		}
	}

	return false
}

// ====================================
// Stage
// ====================================
func (c *ShipRestsController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Dettaglio riposo
	// ##################################################################################################
	case 0:
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

		// Recupero orario per riprendere attività
		var playAt time.Time
		if playAt, err = helpers.GetEndTime(rStartPlayerRest.GetPlayEndTime(), c.Player); err != nil {
			c.Logger.Panic(err)
		}

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
			// Verifico quanto è necessario riposare per tornare in vita/svolgere le attività
			if restsInfo.GetPlayRestsTime() > 0 {
				restsRecap += helpers.Trans(c.Player.Language.Slug, "ship.rests.play_rest_time", restsInfo.GetPlayRestsTime(), playAt.Format("15:04:05"))
			}

			restsRecap += helpers.Trans(c.Player.Language.Slug, "ship.rests.full_rest_time", restsInfo.GetFullRestsTime(), finishAt.Format("15:04:05"))
		}

		restsRecap += helpers.Trans(c.Player.Language.Slug, "ship.rests.info")

		// Aggiongo bottone start riposo
		var keyboardRow [][]tgbotapi.KeyboardButton
		if restsInfo.NeedRests {
			newKeyboardRow := tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "ship.rests.wakeup")),
			)
			keyboardRow = append(keyboardRow, newKeyboardRow)
		}

		// Invio messaggio
		msg := helpers.NewMessage(c.Player.ChatID, restsRecap)
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRow,
		}
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 1

	// ##################################################################################################
	// Fine riposo
	// ##################################################################################################
	case 1:
		var rEndPlayerRest *pb.EndPlayerRestResponse
		if rEndPlayerRest, err = config.App.Server.Connection.EndPlayerRest(helpers.NewContext(1), &pb.EndPlayerRestRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Recap Rest
		var recapMessage = helpers.Trans(c.Player.Language.Slug, "ship.rests.need_to_rest")
		if rEndPlayerRest.GetLifeRecovered() > 0 {
			lifeRecovered := rEndPlayerRest.GetLifeRecovered()

			// Verifico overflow vita
			if lifeRecovered > c.Player.GetLevel().GetPlayerMaxLife() {
				lifeRecovered = c.Player.GetLevel().GetPlayerMaxLife()
			}

			recapMessage = helpers.Trans(c.Player.Language.Slug, "ship.rests.finish", lifeRecovered)
		}

		// Invio messaggio
		msg := helpers.NewMessage(c.ChatID, recapMessage)
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "route.breaker.menu"),
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
