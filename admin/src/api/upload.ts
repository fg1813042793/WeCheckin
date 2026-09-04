import type { UploadRequestOptions } from 'element-plus'
import { ElMessage } from 'element-plus'
import request from '../utils/request'
import { hasPerm } from '../utils/permission'

export const ADMIN_UPLOAD_PERMISSION = 'admin:api:upload:create'

export interface UploadResult {
  url: string
  thumb: string
  path: string
  filename: string
  domain: string
}

export function canUploadAdminFile() {
  return hasPerm(ADMIN_UPLOAD_PERMISSION)
}

export function uploadAdminFile(file: File) {
  if (!canUploadAdminFile()) {
    ElMessage.warning('当前账号没有文件上传权限')
    return Promise.reject(new Error('missing admin upload permission'))
  }
  const form = new FormData()
  form.append('file', file)
  return request.post<UploadResult, FormData>('/api/v2/admin/uploads', form)
}

export async function adminUploadRequest(options: UploadRequestOptions) {
  try {
    const response = await uploadAdminFile(options.file)
    options.onSuccess(response)
  } catch (error) {
    const uploadError = Object.assign(new Error('上传失败'), {
      name: 'UploadAjaxError',
      status: 0,
      method: 'POST',
      url: '/api/v2/admin/uploads',
      cause: error,
    })
    options.onError(uploadError)
  }
}
