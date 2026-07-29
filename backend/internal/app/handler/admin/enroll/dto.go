package enroll

import "wecheckin-backend/backend/internal/model"

type enrollListItem struct {
	ID             uint   `json:"id"`
	Title          string `json:"title"`
	Description    string `json:"desc"`
	Image          string `json:"img"`
	Status         int    `json:"status"`
	DeptID         uint   `json:"deptId"`
	PublishDeptIds string `json:"publishDeptIds"`
	CreateBy       uint   `json:"createBy"`
	CateID         string `json:"cateId"`
	CateName       string `json:"cateName"`
	Start          int64  `json:"timeStart"`
	End            int64  `json:"timeEnd"`
	DayCnt         int    `json:"dayCnt"`
	Order          int    `json:"order"`
	Vouch          int    `json:"vouch"`
	AllowRepeat    bool   `json:"allowRepeat"`
	DailyLimit     int    `json:"dailyLimit"`
	QR             string `json:"qr"`
	ViewCnt        int    `json:"viewCnt"`
	JoinCnt        int    `json:"joinCount"`
	UserCnt        int    `json:"userCnt"`
	AddTime        int64  `json:"_createTime"`
	EditTime       int64  `json:"editTime"`
	StartStr       string `json:"start"`
	EndStr         string `json:"end"`
	StatusDesc     string `json:"statusDesc"`
}

type pagedListResponse struct {
	List  []enrollListItem `json:"list"`
	Total int64            `json:"total"`
}

type enrollJoinListResponse struct {
	List  []model.EnrollJoin `json:"list"`
	Total int64              `json:"total"`
}

func newEnrollListItems(list []model.Enroll) []enrollListItem {
	result := make([]enrollListItem, 0, len(list))
	for _, item := range list {
		result = append(result, enrollListItem{
			ID:             item.ID,
			Title:          item.Title,
			Description:    item.Desc,
			Image:          item.Img,
			Status:         item.Status,
			DeptID:         item.DeptID,
			PublishDeptIds: item.PublishDeptIds,
			CreateBy:       item.CreateBy,
			CateID:         item.CateID,
			CateName:       item.CateName,
			Start:          item.Start,
			End:            item.End,
			DayCnt:         item.DayCnt,
			Order:          item.Order,
			Vouch:          item.Vouch,
			AllowRepeat:    item.AllowRepeat,
			DailyLimit:     item.DailyLimit,
			QR:             item.QR,
			ViewCnt:        item.ViewCnt,
			JoinCnt:        item.JoinCnt,
			UserCnt:        item.UserCnt,
			AddTime:        item.AddTime,
			EditTime:       item.EditTime,
			StartStr:       item.StartStr,
			EndStr:         item.EndStr,
			StatusDesc:     item.StatusDesc,
		})
	}
	return result
}
