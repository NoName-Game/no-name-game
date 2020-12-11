package helpers

// GetResourceCategoryIcons
func GetResourceCategoryIcons(categoryID uint32) (category string) {
	switch categoryID {
	case 1:
		category = "🔥"
	case 2:
		category = "💧"
	case 3:
		category = "⚡️"
	}
	return
}

// GetResourceBaseIcons
func GetResourceBaseIcons(isBase bool) (result string) {
	if isBase {
		result = "🔬Base"
	}
	return
}
