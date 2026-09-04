import { get, post, put, patch, del } from '../utils/request'

const API_V2 = '/api/v2'
const ADMIN_V2 = `${API_V2}/admin`

function valueOf(source, keys = ['id'], fallback = '') {
  if (source === null || source === undefined) return fallback
  if (typeof source !== 'object') return source
  for (const key of keys) {
    const value = source[key]
    if (value !== undefined && value !== null && value !== '') return value
  }
  return fallback
}

function pathParam(source, keys = ['id'], fallback = '') {
  return encodeURIComponent(String(valueOf(source, keys, fallback)))
}

export const dictApi = {
  items(typeCode) {
    return get(`${API_V2}/dict/items`, { typeCode })
  }
}

export const adminApi = {
  home() {
    return get(`${ADMIN_V2}/home`)
  },
  login(data) {
    return post(`${ADMIN_V2}/auth/login`, data)
  },
  userList(params) {
    return get(`${ADMIN_V2}/users`, params)
  },
  userDetail(openid) {
    return get(`${ADMIN_V2}/users/by-openid/${pathParam(openid)}`, { openid })
  },
  userDetailById(id) {
    return get(`${ADMIN_V2}/users/${pathParam(id)}`)
  },
  userAdd(data) {
    return post(`${ADMIN_V2}/users`, data)
  },
  userEdit(data) {
    return put(`${ADMIN_V2}/users/${pathParam(data)}`, data)
  },
  userStatus(data) {
    return patch(`${ADMIN_V2}/users/${pathParam(data)}/status`, data)
  },
  userDel(data) {
    return del(`${ADMIN_V2}/users/${pathParam(data)}`, data)
  },
  userDataExport(params) {
    return get(`${ADMIN_V2}/users/data/export`, params)
  },
  userDataGet(params) {
    return get(`${ADMIN_V2}/users/data`, params)
  },
  userDataDel(params) {
    return del(`${ADMIN_V2}/users/data/${pathParam(params, ['id'], 'current')}`, params)
  },
  userFormFields() {
    return get(`${ADMIN_V2}/users/form-fields`)
  },
  userFormFieldSave(data) {
    return put(`${ADMIN_V2}/users/form-fields`, data)
  },
  enrollList(params) {
    return get(`${ADMIN_V2}/enrollments`, params)
  },
  enrollDetail(id) {
    return get(`${ADMIN_V2}/enrollments/${pathParam(id)}`)
  },
  enrollInsert(data) {
    return post(`${ADMIN_V2}/enrollments`, data)
  },
  enrollEdit(data) {
    return put(`${ADMIN_V2}/enrollments/${pathParam(data)}`, data)
  },
  enrollDel(data) {
    return del(`${ADMIN_V2}/enrollments/${pathParam(data)}`, data)
  },
  enrollStatus(data) {
    return patch(`${ADMIN_V2}/enrollments/${pathParam(data)}/status`, data)
  },
  enrollSort(data) {
    return patch(`${ADMIN_V2}/enrollments/${pathParam(data)}/sort`, data)
  },
  enrollVouch(data) {
    return patch(`${ADMIN_V2}/enrollments/${pathParam(data)}/recommendation`, data)
  },
  enrollClear(data) {
    return post(`${ADMIN_V2}/enrollments/${pathParam(data)}/clear`, data)
  },
  enrollJoinList(params) {
    return get(`${ADMIN_V2}/enrollments/${pathParam(params, ['enrollId', 'id'])}/joins`, params)
  },
  enrollUserList(params) {
    return get(`${ADMIN_V2}/enrollments/${pathParam(params, ['enrollId', 'id'])}/users`, params)
  },
  enrollRemoveUser(data) {
    return del(`${ADMIN_V2}/enrollments/${pathParam(data, ['enrollId', 'id'])}/users/${pathParam(data, ['userId'])}`, data)
  },
  enrollUserFormsEdit(data) {
    return put(`${ADMIN_V2}/enrollments/${pathParam(data, ['enrollId', 'id'])}/users/${pathParam(data, ['userId'])}/forms`, data)
  },
  enrollJoinDel(data) {
    return del(`${ADMIN_V2}/enrollments/${pathParam(data, ['enrollId', 'id'], 0)}/joins/${pathParam(data, ['enrollJoinId', 'joinId', 'id'])}`, data)
  },
  enrollJoinDataExport(params) {
    return post(`${ADMIN_V2}/enrollments/${pathParam(params, ['enrollId', 'id'])}/export`, params)
  },
  enrollJoinDataGet(params) {
    return get(`${ADMIN_V2}/enrollments/${pathParam(params, ['enrollId', 'id'])}/export`, params)
  },
  enrollJoinDataDel(params) {
    return del(`${ADMIN_V2}/enrollments/${pathParam(params, ['enrollId', 'id'])}/export`, params)
  },
  newsList(params) {
    return get(`${ADMIN_V2}/news`, params)
  },
  newsDetail(id) {
    return get(`${ADMIN_V2}/news/${pathParam(id)}`)
  },
  newsInsert(data) {
    return post(`${ADMIN_V2}/news`, data)
  },
  newsEdit(data) {
    return put(`${ADMIN_V2}/news/${pathParam(data)}`, data)
  },
  newsDel(data) {
    return del(`${ADMIN_V2}/news/${pathParam(data)}`, data)
  },
  newsStatus(data) {
    return patch(`${ADMIN_V2}/news/${pathParam(data)}/status`, data)
  },
  newsSort(data) {
    return patch(`${ADMIN_V2}/news/${pathParam(data)}/sort`, data)
  },
  newsVouch(data) {
    return patch(`${ADMIN_V2}/news/${pathParam(data)}/recommendation`, data)
  },
  mgrList(params) {
    return get(`${ADMIN_V2}/managers`, params)
  },
  mgrDetail(id) {
    return get(`${ADMIN_V2}/managers/${pathParam(id)}`)
  },
  mgrInsert(data) {
    return post(`${ADMIN_V2}/managers`, data)
  },
  mgrEdit(data) {
    return put(`${ADMIN_V2}/managers/${pathParam(data)}`, data)
  },
  mgrDel(data) {
    return del(`${ADMIN_V2}/managers/${pathParam(data)}`, data)
  },
  mgrStatus(data) {
    return patch(`${ADMIN_V2}/managers/${pathParam(data)}/status`, data)
  },
  mgrPwd(data) {
    return patch(`${ADMIN_V2}/managers/${pathParam(data)}/password`, data)
  },
  logList(params) {
    return get(`${ADMIN_V2}/logs`, params)
  },
  logClear(data) {
    return del(`${ADMIN_V2}/logs`, data)
  },
  setupQr(params) {
    return get(`${ADMIN_V2}/settings/mini-qr`, params)
  },
  setupSet(data) {
    return put(`${ADMIN_V2}/settings`, data)
  },
  setupSetContent(data) {
    return put(`${ADMIN_V2}/settings/content`, data)
  },
  clearVouch(data) {
    return del(`${ADMIN_V2}/home/recommendations`, data)
  },
  eventList(params) {
    return get(`${ADMIN_V2}/events`, params)
  },
  eventDetail(id) {
    return get(`${ADMIN_V2}/events/${pathParam(id)}`)
  },
  eventInsert(data) {
    return post(`${ADMIN_V2}/events`, data)
  },
  eventEdit(data) {
    return put(`${ADMIN_V2}/events/${pathParam(data)}`, data)
  },
  eventDel(data) {
    return del(`${ADMIN_V2}/events/${pathParam(data)}`, data)
  },
  eventStatus(data) {
    return patch(`${ADMIN_V2}/events/${pathParam(data)}/status`, data)
  },
  eventVouch(data) {
    return patch(`${ADMIN_V2}/events/${pathParam(data)}/recommendation`, data)
  },
  eventTop(data) {
    return patch(`${ADMIN_V2}/events/${pathParam(data)}/top`, data)
  },
  eventParticipantList(params) {
    return get(`${ADMIN_V2}/events/${pathParam(params, ['eventId', 'id'])}/participants`, params)
  },
  eventParticipantDel(data) {
    return del(`${ADMIN_V2}/events/${pathParam(data, ['eventId'], 0)}/participants/${pathParam(data)}`, data)
  },
  eventParticipantEdit(data) {
    return put(`${ADMIN_V2}/events/${pathParam(data, ['eventId'], 0)}/participants/${pathParam(data)}`, data)
  },
  deptUsers(params) {
    return get(`${ADMIN_V2}/event-dept-users`, params)
  },
  deptTree() {
    return get(`${ADMIN_V2}/departments/tree`)
  },
  roleList(params) {
    return get(`${ADMIN_V2}/roles`, params)
  },
  adminPerms() {
    return get(`${ADMIN_V2}/me/perms`)
  },
  formkitTypes() {
    return get(`${ADMIN_V2}/survey-types`)
  },
  formkitParseSchema(data) {
    return post(`${ADMIN_V2}/survey-schema/parse`, data)
  }
}
