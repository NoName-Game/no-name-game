package helpers

import (
	"fmt"
	"reflect"
	"strings"

	"bitbucket.org/no-name-game/no-name/app/acme/nnsdk"
	"bitbucket.org/no-name-game/no-name/services"
)

// PlayerStatsToString - Convert player stats to string
func PlayerStatsToString(playerStats *nnsdk.PlayerStats, playerLanguageSlug string) (result string) {
	val := reflect.ValueOf(playerStats).Elem()
	for i := 4; i < val.NumField()-1; i++ {
		valueField := val.Field(i)
		fieldName, _ := services.GetTranslation("ability."+strings.ToLower(val.Type().Field(i).Name), playerLanguageSlug, nil)

		result += fmt.Sprintf("<code>%-15v:%v</code>\n", fieldName, valueField.Interface())
	}
	return
}

// Increment - Increment player stats by fieldName
func PlayerStatsIncrement(playerStats *nnsdk.PlayerStats, statToIncrement string, playerLanguageSlug string) {
	val := reflect.ValueOf(playerStats).Elem()
	for i := 3; i < val.NumField()-1; i++ {
		fieldName, _ := services.GetTranslation("ability."+strings.ToLower(val.Type().Field(i).Name), playerLanguageSlug, nil)

		if fieldName == statToIncrement {
			f := reflect.ValueOf(playerStats).Elem().FieldByName(val.Type().Field(i).Name)
			f.SetUint(uint64(f.Interface().(uint) + 1))
			playerStats.AbilityPoint--
		}
	}
}
