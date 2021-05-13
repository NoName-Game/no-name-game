package helpers

func GetShipCategoryIcons(categoryID uint32) (category string) {
	switch categoryID {
	case 1:
		category = "💨"
	case 2:
		category = "🥊"
	case 3:
		category = "🛡"
	}
	return
}
// ▰ ▱
func GenerateHealthBar(integrity uint32) string {
	result := ""
	for i := uint32(0); i < 10; i++ {
		if i < integrity/10 {
			result += "▰"
		} else {
			result += "▱"
		}
	}

	return result
}
