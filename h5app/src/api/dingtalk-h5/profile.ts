import type { AvatarUploadPayload, ChangePasswordRequest, UpdateProfileRequest } from '@/types/dingtalk-h5'
import { DINGTALK_H5_API, patch, uploadFile } from './base'

export function uploadAvatar(filePath: string) {
  return uploadFile<AvatarUploadPayload>(`${DINGTALK_H5_API}/account/avatar`, filePath)
}

export function updateProfile(data: UpdateProfileRequest) {
  return patch(`${DINGTALK_H5_API}/account/profile`, data)
}

export function changePassword(data: ChangePasswordRequest) {
  return patch(`${DINGTALK_H5_API}/account/password`, data)
}
