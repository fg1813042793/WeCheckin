package dept

import (
	"context"

	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

// UserDeptID returns the first department associated with a user.
func UserDeptID(userID uint) uint {
	return UserDeptIDContext(context.Background(), userID)
}

func UserDeptIDContext(ctx context.Context, userID uint) uint {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	var ud model.UserDept
	db.Where("`user_dept_user_id` = ?", userID).First(&ud)
	return ud.DeptID
}

// UserDeptIDs returns all departments associated with a user.
func UserDeptIDs(userID uint) []uint {
	return UserDeptIDsContext(context.Background(), userID)
}

func UserDeptIDsContext(ctx context.Context, userID uint) []uint {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	var depts []model.UserDept
	db.Where("`user_dept_user_id` = ?", userID).Find(&depts)
	ids := make([]uint, 0, len(depts))
	for _, d := range depts {
		ids = append(ids, d.DeptID)
	}
	return ids
}

// UserDeptIDsByMiniOpenID returns user departments and ancestors for publish-scope checks.
func UserDeptIDsByMiniOpenID(miniOpenID string) []uint {
	return UserDeptIDsByMiniOpenIDContext(context.Background(), miniOpenID)
}

func UserDeptIDsByMiniOpenIDContext(ctx context.Context, miniOpenID string) []uint {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	var user model.User
	if err := db.Where("`user_mini_openid` = ?", miniOpenID).First(&user).Error; err != nil {
		return nil
	}
	ids := UserDeptIDsContext(ctx, user.ID)
	seen := map[uint]bool{}
	for _, id := range ids {
		seen[id] = true
	}
	for _, id := range ids {
		for _, aid := range AncestorIDsContext(ctx, id) {
			if !seen[aid] {
				ids = append(ids, aid)
				seen[aid] = true
			}
		}
	}
	return ids
}

// AncestorIDs returns all parent department IDs from child to root.
func AncestorIDs(deptID uint) []uint {
	return AncestorIDsContext(context.Background(), deptID)
}

func AncestorIDsContext(ctx context.Context, deptID uint) []uint {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	var result []uint
	visited := map[uint]bool{}
	for deptID > 0 {
		if visited[deptID] {
			break
		}
		visited[deptID] = true
		var dept model.Department
		if err := db.First(&dept, deptID).Error; err != nil {
			break
		}
		if dept.ParentID > 0 {
			result = append(result, dept.ParentID)
		}
		deptID = dept.ParentID
	}
	return result
}

// TopDeptName walks department ancestors and returns the root department name.
func TopDeptName(deptID uint) string {
	return TopDeptNameContext(context.Background(), deptID)
}

func TopDeptNameContext(ctx context.Context, deptID uint) string {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	visited := map[uint]bool{}
	for deptID > 0 {
		if visited[deptID] {
			break
		}
		visited[deptID] = true
		var dept model.Department
		if err := db.First(&dept, deptID).Error; err != nil {
			break
		}
		if dept.ParentID == 0 {
			return dept.Name
		}
		deptID = dept.ParentID
	}
	return ""
}
