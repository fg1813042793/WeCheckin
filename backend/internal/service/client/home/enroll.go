package home

import (
	"encoding/json"

	"wecheckin/backend/internal/support/media"
	"wecheckin/backend/internal/model"
)

type enrollObj struct {
	Cover []string `json:"cover"`
	Desc  string   `json:"desc"`
}

func populateEnrollFields(list []model.Enroll) []model.Enroll {
	for i := range list {
		var obj enrollObj
		if list[i].Obj != "" {
			json.Unmarshal([]byte(list[i].Obj), &obj)
		}
		if len(obj.Cover) > 0 {
			list[i].Img = media.FullURLWithStaticDomain(obj.Cover[0])
		}
		list[i].Desc = obj.Desc

		if list[i].UserList != "" {
			var arr []map[string]interface{}
			if err := json.Unmarshal([]byte(list[i].UserList), &arr); err == nil {
				list[i].UserListArr = arr
			}
		}
	}
	return list
}
