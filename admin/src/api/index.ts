import request from '../utils/request'

export const adminApi = {
  login(data: { name: string; password: string }) {
    return request.post('/admin/login', data)
  },
  home() {
    return request.get('/admin/home')
  },
  clearVouch() {
    return request.post('/admin/clear_vouch')
  },
  // 用户管理
  userList(params?: any) {
    return request.get('/admin/user_list', { params })
  },
  userDetailById(id: string | number) {
    return request.get('/admin/user_detail_by_id', { params: { id } })
  },
  userAdd(data: any) {
    return request.post('/admin/user_add', data)
  },
  userEdit(data: any) {
    return request.post('/admin/user_edit', data)
  },
  userStatus(data: any) {
    return request.post('/admin/user_status', data)
  },
  userDel(data: any) {
    return request.post('/admin/user_del', data)
  },
  userDels(data: any) {
    return request.post('/admin/user_dels', data)
  },
  userResetPwd(data: any) {
    return request.post('/admin/user_reset_pwd', data)
  },
  userFormFields() {
    return request.get('/user_form_fields')
  },
  userFormFieldSave(data: any) {
    return request.post('/admin/user_form_field_save', data)
  },
  // 打卡管理
  enrollList(params?: any) {
    return request.get('/admin/enroll_list', { params })
  },
  enrollDetail(id: string | number) {
    return request.get('/admin/enroll_detail', { params: { id } })
  },
  enrollInsert(data: any) {
    return request.post('/admin/enroll_insert', data)
  },
  enrollEdit(data: any) {
    return request.post('/admin/enroll_edit', data)
  },
  enrollDel(data: any) {
    return request.post('/admin/enroll_del', data)
  },
  enrollDels(data: any) {
    return request.post('/admin/enroll_dels', data)
  },
  enrollStatus(data: any) {
    return request.post('/admin/enroll_status', data)
  },
  enrollSort(data: any) {
    return request.post('/admin/enroll_sort', data)
  },
  enrollVouch(data: any) {
    return request.post('/admin/enroll_vouch', data)
  },
  enrollClear(data: any) {
    return request.post('/admin/enroll_clear', data)
  },
  enrollJoinList(params?: any) {
    return request.get('/admin/enroll_join_list', { params })
  },
  enrollUserList(params?: any) {
    return request.get('/admin/enroll_user_list', { params })
  },
  enrollStats(params?: any) {
    return request.get('/admin/enroll_stats', { params })
  },
  enrollRemoveUser(data: any) {
    return request.post('/admin/enroll_remove_user', data)
  },
  enrollRemoveUsers(data: any) {
    return request.post('/admin/enroll_remove_users', data)
  },
  enrollJoinDel(data: any) {
    return request.post('/admin/enroll_join_del', data)
  },
  enrollJoinDels(data: any) {
    return request.post('/admin/enroll_join_dels', data)
  },
  enrollJoinDataExport(params?: any) {
    return request.get('/admin/enroll_join_data_export', { params })
  },
  // 内容管理
  newsList(params?: any) {
    return request.get('/admin/news_list', { params })
  },
  newsDetail(id: string | number) {
    return request.get('/admin/news_detail', { params: { id } })
  },
  newsInsert(data: any) {
    return request.post('/admin/news_insert', data)
  },
  newsEdit(data: any) {
    return request.post('/admin/news_edit', data)
  },
  newsDel(data: any) {
    return request.post('/admin/news_del', data)
  },
  newsDels(data: any) {
    return request.post('/admin/news_dels', data)
  },
  newsStatus(data: any) {
    return request.post('/admin/news_status', data)
  },
  newsVouch(data: any) {
    return request.post('/admin/news_vouch', data)
  },
  newsSort(data: any) {
    return request.post('/admin/news_sort', data)
  },
  // 管理员管理
  mgrList(params?: any) {
    return request.get('/admin/mgr_list', { params })
  },
  mgrDetail(id: string | number) {
    return request.get('/admin/mgr_detail', { params: { id } })
  },
  mgrInsert(data: any) {
    return request.post('/admin/mgr_insert', data)
  },
  mgrEdit(data: any) {
    return request.post('/admin/mgr_edit', data)
  },
  mgrDel(data: any) {
    return request.post('/admin/mgr_del', data)
  },
  mgrDels(data: any) {
    return request.post('/admin/mgr_dels', data)
  },
  mgrStatus(data: any) {
    return request.post('/admin/mgr_status', data)
  },
  mgrPwd(data: any) {
    return request.post('/admin/mgr_pwd', data)
  },
  // 操作日志
  logList(params?: any) {
    return request.get('/admin/log_list', { params })
  },
  logClear() {
    return request.post('/admin/log_clear')
  },
  // 设置
  setupSetContent(data: any) {
    return request.post('/admin/setup_set_content', data)
  },
  // 字典管理
  dictTypes() {
    return request.get('/admin/dict/types')
  },
  dictItems(typeCode: string) {
    return request.get('/admin/dict/items', { params: { typeCode } })
  },
  dictAdd(data: any) {
    return request.post('/admin/dict/add', data)
  },
  dictEdit(data: any) {
    return request.post('/admin/dict/edit', data)
  },
  dictDel(data: any) {
    return request.post('/admin/dict/del', data)
  },
  dictClear(typeCode: string) {
    return request.post('/admin/dict/clear', { typeCode })
  },
  dictEditTypeName(data: any) {
    return request.post('/admin/dict/edit_type_name', data)
  },
  // 部门管理
  deptTree() {
    return request.get('/admin/dept/tree')
  },
  deptAdd(data: any) {
    return request.post('/admin/dept/add', data)
  },
  deptEdit(data: any) {
    return request.post('/admin/dept/edit', data)
  },
  deptDel(data: any) {
    return request.post('/admin/dept/del', data)
  },
  // 角色管理
  roleList(params?: any) {
    return request.get('/admin/role/list', { params })
  },
  roleAdd(data: any) {
    return request.post('/admin/role/add', data)
  },
  roleEdit(data: any) {
    return request.post('/admin/role/edit', data)
  },
  roleDel(data: any) {
    return request.post('/admin/role/del', data)
  },
  roleDels(data: any) {
    return request.post('/admin/role/dels', data)
  },
  // 菜单管理
  menuTree() {
    return request.get('/admin/menu/tree')
  },
  menuList() {
    return request.get('/admin/menu/list')
  },
  menuAdd(data: any) {
    return request.post('/admin/menu/add', data)
  },
  menuEdit(data: any) {
    return request.post('/admin/menu/edit', data)
  },
  menuDel(data: any) {
    return request.post('/admin/menu/del', data)
  },
  // 赛事活动管理
  eventList(params?: any) {
    return request.get('/admin/event_list', { params })
  },
  eventDetail(id: string | number) {
    return request.get('/admin/event_detail', { params: { id } })
  },
  eventInsert(data: any) {
    return request.post('/admin/event_insert', data)
  },
  eventEdit(data: any) {
    return request.post('/admin/event_edit', data)
  },
  eventDel(data: any) {
    return request.post('/admin/event_del', data)
  },
  eventDels(data: any) {
    return request.post('/admin/event_dels', data)
  },
  eventStatus(data: any) {
    return request.post('/admin/event_status', data)
  },
  eventParticipantList(params?: any) {
    return request.get('/admin/event_participant_list', { params })
  },
  eventParticipantDel(data: any) {
    return request.post('/admin/event_participant_del', data)
  },
  eventParticipantDels(data: any) {
    return request.post('/admin/event_participant_dels', data)
  },
  eventDynamics(params?: any) {
    return request.get('/admin/event_dynamics', { params })
  },
  eventDynamicAdd(data: any) {
    return request.post('/admin/event_dynamic_add', data)
  },
  eventDynamicEdit(data: any) {
    return request.post('/admin/event_dynamic_edit', data)
  },
  eventDynamicDel(data: any) {
    return request.post('/admin/event_dynamic_del', data)
  },
  eventDynamicDels(data: any) {
    return request.post('/admin/event_dynamic_dels', data)
  },
  eventScores(params?: any) {
    return request.get('/admin/event_scores', { params })
  },
  eventScoreEdit(data: any) {
    return request.post('/admin/event_score_edit', data)
  },
  eventVouch(data: any) {
    return request.post('/admin/event_vouch', data)
  },
  eventTop(data: any) {
    return request.post('/admin/event_top', data)
  },
  deptUsers(params?: any) {
    return request.get('/admin/dept_users', { params })
  },
  // 当前管理员的菜单和权限
  adminMenus() {
    return request.get('/admin/user/menus')
  },
  adminPerms() {
    return request.get('/admin/user/perms')
  },
  // 在线用户
  onlineUsers() {
    return request.get('/admin/user/online')
  },
  forceOfflineUser(data: { id: string | number }) {
    return request.post('/admin/user/force_offline', data)
  },
  onlineAdmins() {
    return request.get('/admin/admin/online')
  },
  forceOfflineAdmin(data: { id: string | number }) {
    return request.post('/admin/admin/force_offline', data)
  },
  adminLogout() {
    return request.post('/admin/admin/logout')
  }
}
