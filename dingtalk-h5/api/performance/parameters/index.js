import { DINGTALK_H5_API, get, put } from '../../common/base'

export function getTemplate() {
  return get(`${DINGTALK_H5_API}/template`)
}

export function saveTemplate(data) {
  return put(`${DINGTALK_H5_API}/template`, data)
}
