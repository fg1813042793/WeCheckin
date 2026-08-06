import { DINGTALK_H5_API, get, post } from '../../common/base'

export function publicConfig() {
  return get(`${DINGTALK_H5_API}/public-config`)
}

export function login(data) {
  return post(`${DINGTALK_H5_API}/login`, data)
}

export function ssoLogin(data) {
  return post(`${DINGTALK_H5_API}/sso-login`, data)
}

export function logout() {
  return post(`${DINGTALK_H5_API}/logout`)
}
