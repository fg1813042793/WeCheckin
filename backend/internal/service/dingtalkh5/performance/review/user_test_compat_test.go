package review

import (
	"context"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	reviewscope "wecheckin/backend/internal/service/dingtalkh5/performance/review/scope"
	usersvc "wecheckin/backend/internal/service/dingtalkh5/performance/user"
	permissionsupport "wecheckin/backend/internal/support/permission"
)

func hydratePerfUser(user *model.DingTalkH5PerfUser) {
	usersvc.HydratePerfUser(user)
}

func sanitizeUserPayload(payload UserPayload, existing *model.DingTalkH5PerfUser) (model.DingTalkH5PerfUser, error) {
	return usersvc.SanitizeUserPayload(usersvc.UserPayload(payload), existing)
}

func encodePerfUserObj(raw string, user model.DingTalkH5PerfUser) string {
	return usersvc.EncodePerfUserObj(raw, user)
}

func visiblePerfUsers(current *model.DingTalkH5PerfUser, users []model.DingTalkH5PerfUser, scope permissionsupport.DataScope) []model.DingTalkH5PerfUser {
	if current == nil || !scope.Ready {
		return nil
	}
	if scope.Mode == 1 {
		return users
	}
	allowed := map[string]struct{}{}
	addAccount := func(account string) {
		account = NormalizeUserID(account)
		if account != "" {
			allowed[account] = struct{}{}
		}
	}
	addAccount(current.Account)
	if scope.Mode != 3 {
		addAccount(current.ManagerAccount)
		addAccount(current.HRBPAccount)
		for _, user := range users {
			if user.Account == current.Account ||
				user.Account == current.ManagerAccount ||
				user.Account == current.HRBPAccount ||
				user.ManagerAccount == current.Account ||
				user.HRBPAccount == current.Account ||
				reviewscope.PerfUserDepartmentScopeMatches(*current, user) {
				addAccount(user.Account)
				addAccount(user.ManagerAccount)
				addAccount(user.HRBPAccount)
			}
		}
	}
	result := make([]model.DingTalkH5PerfUser, 0, len(allowed))
	for _, user := range users {
		if _, ok := allowed[NormalizeUserID(user.Account)]; ok {
			result = append(result, user)
		}
	}
	return result
}

func dataScopeUserAccountsContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser, scope permissionsupport.DataScope) (map[string]struct{}, bool, error) {
	return usersvc.DataScopeUserAccountsContext(ctx, db, user, scope)
}
