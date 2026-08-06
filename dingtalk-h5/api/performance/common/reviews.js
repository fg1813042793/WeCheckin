import { authToken, buildApiUrl, del, DINGTALK_H5_API, get, post } from '../../common/base'

export function listReviews(params = {}) {
  return get(`${DINGTALK_H5_API}/reviews`, params)
}

export function reviewDetail(id) {
  return get(`${DINGTALK_H5_API}/reviews/${encodeURIComponent(id)}`)
}

export function reviewAction(id, action, data = {}) {
  return post(`${DINGTALK_H5_API}/reviews/${encodeURIComponent(id)}/${action}`, data)
}

export function deleteReview(id) {
  return del(`${DINGTALK_H5_API}/reviews/${encodeURIComponent(id)}`)
}

export function exportReviewsUrl(params = {}) {
  return buildApiUrl(`${DINGTALK_H5_API}/reviews/export`, {
    ...params,
    token: authToken()
  })
}
