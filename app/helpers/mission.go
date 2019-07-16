package helpers

func GetMissionCategoryID(eType string) uint {
	switch eType {
	case Trans("mission.underground"):
		return 2
	case Trans("mission.surface"):
		return 1
	case Trans("mission.atmosphere"):
		return 3
	}
	return 0
}
