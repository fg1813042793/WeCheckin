import { DINGTALK_H5_API, get } from '../../common/base'

export function bootstrap() {
  return get(`${DINGTALK_H5_API}/bootstrap`)
}

export function workbench() {
  return get(`${DINGTALK_H5_API}/workbench`)
}
