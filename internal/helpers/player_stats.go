package helpers

// PlayerStatsToString - Convert player stats to string
// func PlayerStatsToString(playerStats *nnsdk.PlayerStats) (result string) {
// 	val := reflect.ValueOf(playerStats).Elem()
// 	for i := 8; i < val.NumField()-1; i++ {
// 		valueField := val.Field(i)
// 		fieldName := Trans("ability." + strings.ToLower(val.Type().Field(i).Name))
//
// 		result += fmt.Sprintf("<code>%-15v:%v</code>\n", fieldName, valueField.Interface())
// 	}
// 	return
// }
//
// Increment - Increment player stats by fieldName
// func PlayerStatsIncrement(playerStats *nnsdk.PlayerStats, statToIncrement string) {
// 	val := reflect.ValueOf(playerStats).Elem()
// 	for i := 8; i < val.NumField()-1; i++ {
// 		fieldName := Trans("ability." + strings.ToLower(val.Type().Field(i).Name))
//
// 		if fieldName == statToIncrement {
// 			f := reflect.ValueOf(playerStats).Elem().FieldByName(val.Type().Field(i).Name)
// 			f.SetUint(uint64(f.Interface().(uint) + 1))
// 			playerStats.AbilityPoint--
// 		}
// 	}
// }

// DecrementLife - Handle the life points
// func DecrementLife(lifePoint uint, stats nnsdk.PlayerStats) nnsdk.PlayerStats {
// 	// MaxLife = 100 + Level * 10
// 	if *stats.LifePoint-lifePoint > 100+stats.Level*10 { // Overflow problem
// 		*stats.LifePoint = 0
// 	} else {
// 		*stats.LifePoint -= lifePoint
// 	}
//
// 	var err error
// 	stats, err = providers.UpdatePlayerStats(stats)
// 	if err != nil {
// 		services.ErrorHandler("Cant update player stats", err)
// 	}
//
// 	return stats
// }
