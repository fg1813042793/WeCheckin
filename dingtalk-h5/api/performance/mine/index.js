import { DINGTALK_H5_API, post } from '../../common/base'

export function createReview(data) {
  return post(`${DINGTALK_H5_API}/reviews`, data)
}
