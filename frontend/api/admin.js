import { get, post } from '../utils/request'

export const dictApi = {
  items(typeCode) {
    return get('/admin/dict/items', { typeCode })
  }
}

export const adminApi = {
  home() {
    return get('/admin/home')
  },
  login(data) {
    return post('/admin/login', data)
  },
  userList(params) {
    return get('/admin/user_list', params)
  },
  userDetail(openid) {
    return get('/admin/user_detail', { openid })
  },
  userDetailById(id) {
    return get('/admin/user_detail_by_id', { id })
  },
  userAdd(data) {
    return post('/admin/user_add', data)
  },
  userEdit(data) {
    return post('/admin/user_edit', data)
  },
  userStatus(data) {
    return post('/admin/user_status', data)
  },
  userDel(data) {
    return post('/admin/user_del', data)
  },
  userDataExport(params) {
    return get('/admin/user_data_export', params)
  },
  userDataGet(params) {
    return get('/admin/user_data_get', params)
  },
  userDataDel(params) {
    return get('/admin/user_data_del', params)
  },
  userFormFields() {
    return get('/user_form_fields')
  },
  userFormFieldSave(data) {
    return post('/admin/user_form_field_save', data)
  },
  enrollList(params) {
    return get('/admin/enroll_list', params)
  },
  enrollDetail(id) {
    return get('/admin/enroll_detail', { id })
  },
  enrollInsert(data) {
    return post('/admin/enroll_insert', data)
  },
  enrollEdit(data) {
    return post('/admin/enroll_edit', data)
  },
  enrollDel(data) {
    return post('/admin/enroll_del', data)
  },
  enrollStatus(data) {
    return post('/admin/enroll_status', data)
  },
  enrollSort(data) {
    return post('/admin/enroll_sort', data)
  },
  enrollVouch(data) {
    return post('/admin/enroll_vouch', data)
  },
  enrollClear(data) {
    return post('/admin/enroll_clear', data)
  },
  enrollJoinList(params) {
    return get('/admin/enroll_join_list', params)
  },
  enrollUserList(params) {
    return get('/admin/enroll_user_list', params)
  },
  enrollRemoveUser(data) {
    return post('/admin/enroll_remove_user', data)
  },
  enrollUserFormsEdit(data) {
    return post('/admin/enroll_user_forms_edit', data)
  },
  enrollJoinDel(data) {
    return post('/admin/enroll_join_del', data)
  },
  enrollJoinDataExport(params) {
    return get('/admin/enroll_join_data_export', params)
  },
  enrollJoinDataGet(params) {
    return get('/admin/enroll_join_data_get', params)
  },
  enrollJoinDataDel(params) {
    return get('/admin/enroll_join_data_del', params)
  },
  newsList(params) {
    return get('/admin/news_list', params)
  },
  newsDetail(id) {
    return get('/admin/news_detail', { id })
  },
  newsInsert(data) {
    return post('/admin/news_insert', data)
  },
  newsEdit(data) {
    return post('/admin/news_edit', data)
  },
  newsDel(data) {
    return post('/admin/news_del', data)
  },
  newsStatus(data) {
    return post('/admin/news_status', data)
  },
  newsSort(data) {
    return post('/admin/news_sort', data)
  },
  newsVouch(data) {
    return post('/admin/news_vouch', data)
  },
  mgrList(params) {
    return get('/admin/mgr_list', params)
  },
  mgrDetail(id) {
    return get('/admin/mgr_detail', { id })
  },
  mgrInsert(data) {
    return post('/admin/mgr_insert', data)
  },
  mgrEdit(data) {
    return post('/admin/mgr_edit', data)
  },
  mgrDel(data) {
    return post('/admin/mgr_del', data)
  },
  mgrStatus(data) {
    return post('/admin/mgr_status', data)
  },
  mgrPwd(data) {
    return post('/admin/mgr_pwd', data)
  },
  logList(params) {
    return get('/admin/log_list', params)
  },
  logClear(data) {
    return post('/admin/log_clear', data)
  },
  setupQr(params) {
    return get('/admin/setup_qr', params)
  },
  setupSet(data) {
    return post('/admin/setup_set', data)
  },
  setupSetContent(data) {
    return post('/admin/setup_set_content', data)
  },
  clearVouch(data) {
    return post('/admin/clear_vouch', data)
  },
  eventList(params) {
    return get('/admin/event_list', params)
  },
  eventDetail(id) {
    return get('/admin/event_detail', { id })
  },
  eventInsert(data) {
    return post('/admin/event_insert', data)
  },
  eventEdit(data) {
    return post('/admin/event_edit', data)
  },
  eventDel(data) {
    return post('/admin/event_del', data)
  },
  eventStatus(data) {
    return post('/admin/event_status', data)
  },
  eventVouch(data) {
    return post('/admin/event_vouch', data)
  },
  eventTop(data) {
    return post('/admin/event_top', data)
  },
  eventParticipantList(params) {
    return get('/admin/event_participant_list', params)
  },
  eventParticipantDel(data) {
    return post('/admin/event_participant_del', data)
  },
  eventParticipantEdit(data) {
    return post('/admin/event_participant_edit', data)
  },
  deptUsers(params) {
    return get('/admin/dept_users', params)
  },
  deptTree() {
    return get('/admin/dept/tree')
  },
  adminPerms() {
    return get('/admin/user/perms')
  },
  formkitTypes() {
    return get('/survey/types')
  },
  formkitParseSchema(data) {
    return post('/survey/schema/parse', data)
  }
}
