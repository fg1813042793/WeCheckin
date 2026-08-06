import { del, DINGTALK_H5_API, get, post, put } from '../../common/base'

export function listUsers() {
  return get(`${DINGTALK_H5_API}/users`)
}

export function createUser(data) {
  return post(`${DINGTALK_H5_API}/users`, data)
}

export function updateUser(id, data) {
  return put(`${DINGTALK_H5_API}/users/${encodeURIComponent(id)}`, data)
}

export function deleteUser(id) {
  return del(`${DINGTALK_H5_API}/users/${encodeURIComponent(id)}`)
}
