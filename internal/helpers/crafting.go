package helpers

import (
	"nn-grpc/build/pb"
	"nn-telegram/config"
)

func CheckAndReturnCategorySlug(locale string, text string) (result string) {
	for _, slug := range GetAllSlugCategories() {
		if text == Trans(locale, slug) {
			return slug
		}
	}
	return ""
}

func GetAllSlugCategories() (result []string) {
	rGetAllArmorCategory, err := config.App.Server.Connection.GetAllArmorCategory(NewContext(1), &pb.GetAllArmorCategoryRequest{})
	if err != nil {
		panic(err)
	}

	for i := 0; i < len(rGetAllArmorCategory.GetArmorCategories()); i++ {
		result = append(result, rGetAllArmorCategory.GetArmorCategories()[i].Slug)
	}

	return
}
