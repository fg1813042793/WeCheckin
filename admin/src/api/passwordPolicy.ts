import request from '../utils/request'
import type { PasswordPolicy } from '../types/passwordPolicy'

const jsonConfig = { headers: { 'Content-Type': 'application/json' } }

export function getPasswordPolicy() {
  return request.get<PasswordPolicy>('/api/v2/admin/password-policy')
}

export function savePasswordPolicy(policy: PasswordPolicy) {
  return request.put<PasswordPolicy, PasswordPolicy>('/api/v2/admin/password-policy', policy, jsonConfig)
}
