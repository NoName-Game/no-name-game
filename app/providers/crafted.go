package providers

// func GetCraftedByID(id uint) (nnsdk.Crafted, error) {
// 	var resource nnsdk.Crafted
// 	resp, err := services.NnSDK.MakeRequest("craftable/"+strconv.FormatUint(uint64(id), 10), nil).Get()
// 	if err != nil {
// 		return resource, err
// 	}
//
// 	err = json.Unmarshal(resp.Data, &resource)
// 	if err != nil {
// 		return resource, err
// 	}
//
// 	return resource, nil
// }
//
// func GetAllCraftableItems() (nnsdk.Crafts, error) {
// 	var crafts nnsdk.Crafts
// 	resp, err := services.NnSDK.MakeRequest("craftable", nil).Get()
//
// 	if err != nil {
// 		return crafts, err
// 	}
//
// 	err = json.Unmarshal(resp.Data, &crafts)
//
// 	if err != nil {
// 		return crafts, err
// 	}
//
// 	return crafts, nil
// }
//
// func GetCraftedByName(name string) (nnsdk.Crafted, error) {
//
// 	var craft nnsdk.Crafted
//
// 	params := url.Values{}
// 	params.Add("name", name)
//
// 	resp, err := services.NnSDK.MakeRequest("search/craft?"+params.Encode(), nil).Get()
//
// 	if err != nil {
// 		return craft, err
// 	}
//
// 	err = json.Unmarshal(resp.Data, &craft)
//
// 	if err != nil {
// 		return craft, err
// 	}
// 	return craft, nil
//
// }
