package enroll

import (
	"encoding/json"

	"wecheckin-backend/backend/internal/app/support/media"
	"wecheckin-backend/backend/internal/model"
)

type enrollObj struct {
	Cover []string `json:"cover"`
	Desc  string   `json:"desc"`
}

func decodeEnrollObj(raw string) enrollObj {
	var obj enrollObj
	if raw != "" {
		_ = json.Unmarshal([]byte(raw), &obj)
	}
	return obj
}

func populateEnrollFields(list []model.Enroll) []model.Enroll {
	for i := range list {
		obj := decodeEnrollObj(list[i].Obj)
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
