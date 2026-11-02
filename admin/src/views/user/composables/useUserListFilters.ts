import { reactive } from 'vue'
import type { AdminUserListQuery } from '../../../api/types'

const emptyUserListFilters = (): AdminUserListQuery => ({
  keyword: '', mobile: '', status: undefined,
  positionId: undefined, managerUserId: undefined, roleId: undefined,
})

export function useUserListFilters(onSearch: () => void) {
  const userListFilters = reactive<AdminUserListQuery>(emptyUserListFilters())

  function updateUserListFilters(value: AdminUserListQuery) {
    Object.assign(userListFilters, value)
  }

  function resetUserFilters() {
    Object.assign(userListFilters, emptyUserListFilters())
    onSearch()
  }

  function buildUserListQuery(page: number, pageSize: number): AdminUserListQuery {
    return {
      page,
      pageSize,
      keyword: userListFilters.keyword?.trim() || undefined,
      mobile: userListFilters.mobile?.trim() || undefined,
      status: userListFilters.status,
      positionId: userListFilters.positionId,
      managerUserId: userListFilters.managerUserId,
      roleId: userListFilters.roleId,
    }
  }

  return { userListFilters, updateUserListFilters, resetUserFilters, buildUserListQuery }
}
