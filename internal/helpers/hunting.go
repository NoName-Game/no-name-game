package helpers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	// Hunting Move Actions
	moveDown = InlineDataStruct{
		C:  "hunting",
		AT: "move",
		A:  "down",
		D:  1,
	}

	moveDownFast = InlineDataStruct{
		C:  "hunting",
		AT: "move",
		A:  "down",
		D:  3,
	}

	moveUp = InlineDataStruct{
		C:  "hunting",
		AT: "move",
		A:  "up",
		D:  1,
	}

	moveUpFast = InlineDataStruct{
		C:  "hunting",
		AT: "move",
		A:  "up",
		D:  3,
	}

	moveLeft = InlineDataStruct{
		C:  "hunting",
		AT: "move",
		A:  "left",
		D:  1,
	}

	moveLeftFast = InlineDataStruct{
		C:  "hunting",
		AT: "move",
		A:  "left",
		D:  3,
	}

	moveRight = InlineDataStruct{
		C:  "hunting",
		AT: "move",
		A:  "right",
		D:  1,
	}

	moveRightFast = InlineDataStruct{
		C:  "hunting",
		AT: "move",
		A:  "right",
		D:  3,
	}

	moveAction = InlineDataStruct{
		C:  "hunting",
		AT: "move",
		A:  "action",
	}

	MoveNoAction = InlineDataStruct{
		C:  "hunting",
		AT: "move",
		A:  "no_action",
	}

	TresureKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬆️", moveUp.GetDataString())),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️", moveLeft.GetDataString()),
			tgbotapi.NewInlineKeyboardButtonData("❓️", moveAction.GetDataString()),
			tgbotapi.NewInlineKeyboardButtonData("➡️", moveRight.GetDataString()),
		),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬇️", moveDown.GetDataString())),
	)

	EnemyKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬆️", moveUp.GetDataString())),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️", moveLeft.GetDataString()),
			tgbotapi.NewInlineKeyboardButtonData("⚔️", FightStart.GetDataString()),
			tgbotapi.NewInlineKeyboardButtonData("➡️", moveRight.GetDataString()),
		),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬇️", moveDown.GetDataString())),
	)

	HuntingFightKeyboard = [][]tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔼", FightUp.GetDataString())),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏃‍♂️💨", FightReturnMap.GetDataString()),
			tgbotapi.NewInlineKeyboardButtonData("⚔️", FightHit.GetDataString()),
		),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔽", FightDown.GetDataString())),
	}
)

func GetMapMovementKeyboard(lastMove string) *tgbotapi.InlineKeyboardMarkup {
	var center tgbotapi.InlineKeyboardButton
	switch lastMove {
	case "up":
		center = tgbotapi.NewInlineKeyboardButtonData("⏫️", moveUpFast.GetDataString())
	case "right":
		center = tgbotapi.NewInlineKeyboardButtonData("⏩", moveRightFast.GetDataString())
	case "left":
		center = tgbotapi.NewInlineKeyboardButtonData("⏪", moveLeftFast.GetDataString())
	case "down":
		center = tgbotapi.NewInlineKeyboardButtonData("⏬️", moveDownFast.GetDataString())
	default:
		center = tgbotapi.NewInlineKeyboardButtonData("⏺️", MoveNoAction.GetDataString())
	}

	// Keyboard inline di esplorazione
	keyaboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬆️", moveUp.GetDataString())),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️", moveLeft.GetDataString()),
			center,
			tgbotapi.NewInlineKeyboardButtonData("➡️", moveRight.GetDataString()),
		),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬇️", moveDown.GetDataString())),
	)

	return &keyaboard
}
