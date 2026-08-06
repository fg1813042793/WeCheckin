import { DINGTALK_H5_API, patch, uploadFile } from '../common/base'

export function uploadAvatar(filePath) {
  return uploadFile(`${DINGTALK_H5_API}/account/avatar`, filePath)
}

export function updateProfile(data) {
  return patch(`${DINGTALK_H5_API}/account/profile`, data)
}

export function changePassword(data) {
  return patch(`${DINGTALK_H5_API}/account/password`, data)
}
