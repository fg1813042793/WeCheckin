import type {
  ApiEnvelope,
  AuthSessionPayload,
  BindSelfRequest,
  LoginRequest,
  PublicConfigPayload,
  SsoLoginRequest,
} from '@/types/dingtalk-h5'
import { DINGTALK_H5_API, get, post } from './base'

export function publicConfig() {
  return get<PublicConfigPayload>(`${DINGTALK_H5_API}/public-config`)
}

export function login(data: LoginRequest) {
  return post<AuthSessionPayload>(`${DINGTALK_H5_API}/login`, data)
}

export function ssoLogin(data: SsoLoginRequest) {
  return post<AuthSessionPayload>(`${DINGTALK_H5_API}/sso-login`, data)
}

export function bindSelf(data: BindSelfRequest) {
  return post<AuthSessionPayload>(`${DINGTALK_H5_API}/bind-self`, data)
}

export function logout() {
  return post<ApiEnvelope>(`${DINGTALK_H5_API}/logout`)
}
