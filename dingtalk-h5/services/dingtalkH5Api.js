import { authToken, buildApiUrl, del, get, patch, post, put } from '../utils/request'

const API_V2 = '/api/v2'
const DINGTALK_H5_API = `${API_V2}/dingtalk/h5`

export const dingTalkAuthApi = {
  login(data) {
    return post(`${DINGTALK_H5_API}/login`, data)
  },
  logout() {
    return post(`${DINGTALK_H5_API}/logout`)
  },
  changePassword(data) {
    return patch(`${DINGTALK_H5_API}/account/password`, data)
  }
}

export const dingTalkPerformanceApi = {
  bootstrap() {
    return get(`${DINGTALK_H5_API}/bootstrap`)
  },
  workbench() {
    return get(`${DINGTALK_H5_API}/workbench`)
  },
  template() {
    return get(`${DINGTALK_H5_API}/template`)
  },
  reviews(params = {}) {
    return get(`${DINGTALK_H5_API}/reviews`, params)
  },
  reviewDetail(id) {
    return get(`${DINGTALK_H5_API}/reviews/${encodeURIComponent(id)}`)
  },
  createReview(data) {
    return post(`${DINGTALK_H5_API}/reviews`, data)
  },
  reviewAction(id, action, data = {}) {
    return post(`${DINGTALK_H5_API}/reviews/${encodeURIComponent(id)}/${action}`, data)
  },
  deleteReview(id) {
    return del(`${DINGTALK_H5_API}/reviews/${encodeURIComponent(id)}`)
  },
  exportUrl(params = {}) {
    return buildApiUrl(`${DINGTALK_H5_API}/reviews/export`, {
      ...params,
      token: authToken()
    })
  },
  users() {
    return get(`${DINGTALK_H5_API}/users`)
  },
  createUser(data) {
    return post(`${DINGTALK_H5_API}/users`, data)
  },
  updateUser(id, data) {
    return put(`${DINGTALK_H5_API}/users/${encodeURIComponent(id)}`, data)
  },
  deleteUser(id) {
    return del(`${DINGTALK_H5_API}/users/${encodeURIComponent(id)}`)
  }
}

export { API_V2, DINGTALK_H5_API }
