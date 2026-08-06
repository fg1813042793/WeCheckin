import { DINGTALK_H5_API, post } from '../../common/base'

export function bindSelf(data) {
  return post(`${DINGTALK_H5_API}/bind-self`, data)
}
