package event

import "wecheckin/backend/internal/model"

type listResponse struct {
	List interface{} `json:"list"`
}

type eventListItem struct {
	ID             uint   `json:"id"`
	Title          string `json:"title"`
	Description    string `json:"desc"`
	Image          string `json:"img"`
	Type           int    `json:"type"`
	Status         int    `json:"status"`
	DeptID         uint   `json:"deptId"`
	PublishDeptIds string `json:"publishDeptIds"`
	CreateBy       uint   `json:"createBy"`
	CateID         string `json:"cateId"`
	CateName       string `json:"cateName"`
	RegStart       int64  `json:"regStart"`
	RegEnd         int64  `json:"regEnd"`
	EventStart     int64  `json:"eventStart"`
	EventEnd       int64  `json:"eventEnd"`
	Order          int    `json:"order"`
	Vouch          int    `json:"vouch"`
	IsTop          int    `json:"isTop"`
	QR             string `json:"qr"`
	ViewCnt        int    `json:"viewCnt"`
	JoinCnt        int    `json:"joinCount"`
	UserCnt        int    `json:"userCnt"`
	AddTime        int64  `json:"_createTime"`
	EditTime       int64  `json:"editTime"`
	RegStartStr    string `json:"regStartStr"`
	RegEndStr      string `json:"regEndStr"`
	EventStartStr  string `json:"eventStartStr"`
	EventEndStr    string `json:"eventEndStr"`
	StatusDesc     string `json:"statusDesc"`
}

type pagedListResponse struct {
	List  []eventListItem `json:"list"`
	Total int64           `json:"total"`
}

func newEventListItems(list []model.Event) []eventListItem {
	result := make([]eventListItem, 0, len(list))
	for _, item := range list {
		result = append(result, eventListItem{
			ID:             item.ID,
			Title:          item.Title,
			Description:    item.Desc,
			Image:          item.Img,
			Type:           item.Type,
			Status:         item.Status,
			DeptID:         item.DeptID,
			PublishDeptIds: item.PublishDeptIds,
			CreateBy:       item.CreateBy,
			CateID:         item.CateID,
			CateName:       item.CateName,
			RegStart:       item.RegStart,
			RegEnd:         item.RegEnd,
			EventStart:     item.EventStart,
			EventEnd:       item.EventEnd,
			Order:          item.Order,
			Vouch:          item.Vouch,
			IsTop:          item.IsTop,
			QR:             item.QR,
			ViewCnt:        item.ViewCnt,
			JoinCnt:        item.JoinCnt,
			UserCnt:        item.UserCnt,
			AddTime:        item.AddTime,
			EditTime:       item.EditTime,
			RegStartStr:    item.RegStartStr,
			RegEndStr:      item.RegEndStr,
			EventStartStr:  item.EventStartStr,
			EventEndStr:    item.EventEndStr,
			StatusDesc:     item.StatusDesc,
		})
	}
	return result
}
