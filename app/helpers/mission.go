package helpers

func GetMissionCategoryID(eType, lang string) uint {
	switch eType {
	case Trans("underground", lang):
		return 2
	case Trans("surface", lang):
		return 1
	case Trans("atmosphere", lang):
		return 3
	}
	return 0
}
