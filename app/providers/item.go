package providers

// Writer: reloonfire
// Starting on: 20/01/2020
// Project: no-name-game

// func GetItemByID(id uint) (nnsdk.Item, error) {
// 	var resource nnsdk.Item
// 	resp, err := services.NnSDK.MakeRequest("items/"+strconv.FormatUint(uint64(id), 10), nil).Get()
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
// func GetAllItems() (nnsdk.Items, error) {
// 	var crafts nnsdk.Items
// 	resp, err := services.NnSDK.MakeRequest("items", nil).Get()
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
// func GetItemByName(name string) (nnsdk.Item, error) {
//
// 	var craft nnsdk.Item
//
// 	params := url.Values{}
// 	params.Add("name", name)
//
// 	resp, err := services.NnSDK.MakeRequest("search/item?"+params.Encode(), nil).Get()
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
