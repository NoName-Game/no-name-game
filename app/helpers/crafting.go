package helpers

import (
	"context"
	"strconv"
	"time"

	pb "bitbucket.org/no-name-game/nn-grpc/rpc"
	"bitbucket.org/no-name-game/nn-telegram/services"
)

// ListRecipe - Metodo che aiuta a recuperare la lista di risore necessarie
// al crafting di un determianto item
func ListRecipe(needed map[uint32]int32) (result string, err error) {
	for resourceID, value := range needed {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		response, err := services.NnSDK.GetResourceByID(ctx, &pb.GetResourceByIDRequest{
			ID: resourceID,
		})
		if err != nil {
			return result, err
		}

		var resource *pb.Resource
		resource = response.GetResource()

		result += resource.Name + " x" + strconv.Itoa(int(value)) + "\n"
	}

	return result, err
}

func CheckAndReturnCategorySlug(locale string, text string) (result string) {
	for _, slug := range GetAllSlugCategories() {
		if text == Trans(locale, slug) {
			return slug
		}
	}
	return ""
}

func GetAllSlugCategories() (result []string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	responseArmor, err := services.NnSDK.GetAllArmorCategory(ctx, &pb.GetAllArmorCategoryRequest{})
	if err != nil {
		panic(err)
	}
	aCategories := responseArmor.GetArmorCategory()

	responseWeapon, err := services.NnSDK.GetAllWeaponCategory(ctx, &pb.GetAllWeaponCategoryRequest{})
	if err != nil {
		panic(err)
	}

	wCategories := responseWeapon.GetWeaponCategories()

	for i := 0; i < len(aCategories); i++ {
		result = append(result, aCategories[i].Slug)
	}
	for i := 0; i < len(wCategories); i++ {
		result = append(result, wCategories[i].Slug)
	}
	return
}
