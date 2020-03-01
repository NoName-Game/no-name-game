package helpers

// GetMissionCategory - Recupera il nome originale del tipo di missione
func GetMissionCategory(locale string, eType string) string {

	//TODO: da rivedere
	switch eType {
	case Trans(locale, "mission.underground"):
		return "underground"
	case Trans(locale, "mission.surface"):
		return "surface"
	case Trans(locale, "mission.atmosphere"):
		return "atmosphere"
	}

	return ""
}
