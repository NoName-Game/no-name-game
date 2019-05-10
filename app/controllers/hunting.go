package controllers

import (
	"encoding/json"
	"math/rand"
	"strings"
	"time"

	"bitbucket.org/no-name-game/no-name/app/acme/nnsdk"
	"bitbucket.org/no-name-game/no-name/app/provider"

	"bitbucket.org/no-name-game/no-name/app/commands"
	"bitbucket.org/no-name-game/no-name/app/helpers"
	"bitbucket.org/no-name-game/no-name/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func Hunting(update tgbotapi.Update) {

	message := update.Message
	routeName := "route.hunting"
	state := helpers.StartAndCreatePlayerState(routeName, helpers.Player)

	type payloadHunting struct {
		MobID uint
		Score int //Number of enemy defeated
	}

	var err error

	var payload payloadHunting
	helpers.UnmarshalPayload(state.Payload, &payload)

	var mob nnsdk.Enemy
	if payload.MobID > 0 {
		mob, err = provider.GetEnemyByID(payload.MobID)
		if err != nil {
			services.ErrorHandler("Cant find enemy", err)
		}
	}

	// Stupid poninter stupid json pff
	t := new(bool)
	*t = true

	//====================================
	// Validator
	//====================================
	validationFlag := false
	validationMessage := helpers.Trans("validationMessage", helpers.Player.Language.Slug)
	switch state.Stage {
	case 1:
		if state.FinishAt.Before(time.Now()) {
			validationFlag = true
		} else {
			validationMessage = helpers.Trans("wait", helpers.Player.Language.Slug, state.FinishAt.Format("15:04:05"))
		}
	case 2:
		if message.Text == helpers.Trans("continue", helpers.Player.Language.Slug) && mob.LifePoint > 0 {
			validationFlag = true
		} else if mob.LifePoint == 0 {
			validationFlag = true

			// Delete the enemy from table
			_, err = provider.DeleteEnemy(mob.ID)
			if err != nil {
				services.ErrorHandler("Cant delete enemy", err)
			}

			state.Stage = 4 //Drop
			state, _ = provider.UpdatePlayerState(state)
		}
	case 3:
		if strings.Contains(message.Text, helpers.Trans("combat.attack_with", helpers.Player.Language.Slug)) {
			validationFlag = true
		}
	case 4:
		if message.Text == helpers.Trans("continue", helpers.Player.Language.Slug) {
			validationFlag = true
			state.Stage = 0
		} else if message.Text == helpers.Trans("nope", helpers.Player.Language.Slug) {
			validationFlag = true
			state.Stage = 5
		}
	}

	if !validationFlag {
		if state.Stage != 0 {
			validatorMsg := services.NewMessage(helpers.Player.ChatID, validationMessage)
			services.SendMessage(validatorMsg)
		}
	}

	//====================================
	// LOGIC FLUX:
	// Searching -> Finding -> Fight [Loop] -> Drop
	//====================================

	// FIGHT SYSTEM : Enemy Card / Choose Attack -> Calculate Damage -> Apply Damage

	//====================================
	// Stage
	//====================================
	switch state.Stage {
	case 0:
		// Set timer
		state.FinishAt = commands.GetEndTime(0, int(5*(payload.Score/3)), 0)
		state.ToNotify = t
		state.Stage = 1

		mob, err = provider.Spawn(nnsdk.Enemy{})
		if err != nil {
			services.ErrorHandler("Cant spawn enemy", err)
		}

		payload.MobID = mob.ID
		payload.Score = 1
		payloadUpdated, _ := json.Marshal(payload)
		state.Payload = string(payloadUpdated)
		state, _ = provider.UpdatePlayerState(state)

		services.SendMessage(services.NewMessage(helpers.Player.ChatID, helpers.Trans("hunting.searching", helpers.Player.Language.Slug, state.FinishAt.Format("15:04:05"))))
	case 1:
		if validationFlag {
			// Enemy found
			state.Stage = 2
			state, _ = provider.UpdatePlayerState(state)
			msg := services.NewMessage(helpers.Player.ChatID, helpers.Trans("hunting.enemy.found", helpers.Player.Language.Slug, mob.Name))
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans("continue", helpers.Player.Language.Slug))))
			services.SendMessage(msg)
		}
	case 2:
		if validationFlag {
			state.Stage = 3
			state, _ = provider.UpdatePlayerState(state)
			msg := services.NewMessage(helpers.Player.ChatID, helpers.Trans("hunting.enemy.card", helpers.Player.Language.Slug, mob.Name, mob.LifePoint, helpers.Player.Stats.LifePoint))
			msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
				ResizeKeyboard: true,
				Keyboard:       helpers.GenerateWeaponKeyboard(helpers.Player),
			}
			services.SendMessage(msg)
		}
	case 3:
		if validationFlag {
			// Calculating damage

			weaponName := strings.SplitN(message.Text, " ", 3)[2]

			var weapon nnsdk.Weapon
			weapon, err = provider.FindWeaponByName(weaponName)
			if err != nil {
				services.ErrorHandler("Cant find weapon", err)
			}

			var playerDamage uint

			switch weapon.WeaponCategory.Slug {
			case "knife":
				// Knife damage
				playerDamage = uint(rand.Int31n(6)+1) + (weapon.RawDamage + ((helpers.Player.Stats.Strength + helpers.Player.Stats.Dexterity) / 2))
			default:
				playerDamage = uint(rand.Int31n(6)+1) + (weapon.RawDamage + ((helpers.Player.Stats.Intelligence + helpers.Player.Stats.Dexterity) / 2))
			}

			mob.LifePoint -= playerDamage

			mob, err = provider.UpdateEnemy(mob)
			if err != nil {
				services.ErrorHandler("Cant update enemy", err)
			}

			var text string
			if mob.LifePoint == 0 {
				text = helpers.Trans("combat.last_hit", helpers.Player.Language.Slug)
			} else {
				mobDamage := uint(rand.Int31n(17) + 1)

				var stats nnsdk.PlayerStats
				stats, err = provider.GetPlayerStats(helpers.Player)
				if err != nil {
					services.ErrorHandler("Cant get player stats", err)
				}

				stats = helpers.DecrementLife(mobDamage, stats)
				if stats.LifePoint == 0 {
					// Player Die
					helpers.DeleteRedisAndDbState(helpers.Player)
					msg := services.NewMessage(helpers.Player.ChatID, helpers.Trans("playerDie", helpers.Player.Language.Slug))
					msg.ParseMode = "HTML"
					services.SendMessage(msg)
				}

				text = helpers.Trans("combat.damage", helpers.Player.Language.Slug, playerDamage, mobDamage)
			}

			state.Stage = 2
			state, _ = provider.UpdatePlayerState(state)

			msg := services.NewMessage(helpers.Player.ChatID, text)
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans("continue", helpers.Player.Language.Slug))))
			services.SendMessage(msg)
		}
	case 4:
		if validationFlag {
			helpers.Player.Stats.Experience++
			_, err = provider.UpdatePlayer(helpers.Player)
			if err != nil {
				services.ErrorHandler("Cant update player", err)
			}

			msg := services.NewMessage(helpers.Player.ChatID, helpers.Trans("hunting.experience_earned", helpers.Player.Language.Slug, 1))
			services.SendMessage(msg)
			msg = services.NewMessage(helpers.Player.ChatID, helpers.Trans("hunting.continue", helpers.Player.Language.Slug))
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans("continue", helpers.Player.Language.Slug)), tgbotapi.NewKeyboardButton(helpers.Trans("nope", helpers.Player.Language.Slug))))
			services.SendMessage(msg)
		}
	case 5:
		if validationFlag {
			//====================================
			// IMPORTANT!
			//====================================
			helpers.FinishAndCompleteState(state, helpers.Player)
			//====================================

			msg := services.NewMessage(message.Chat.ID, helpers.Trans("hunting.complete", helpers.Player.Language.Slug))
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("back"),
				),
			)
			services.SendMessage(msg)
		}
	}
}
