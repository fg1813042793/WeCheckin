import { get, post } from '../utils/request'

export const homeApi = {
  getList() {
    return get('/home/list')
  }
}

export const enrollApi = {
  getList() {
    return get('/enroll/list')
  },
  detail(id) {
    return get('/enroll/view', { id })
  },
  join(data) {
    return post('/enroll/join', data)
  },
  myJoinList(params) {
    return get('/enroll/my_join_list', params)
  },
  myUserList(params) {
    return get('/enroll/my_user_list', params)
  }
}

export const newsApi = {
  getList() {
    return get('/news/list')
  },
  detail(id) {
    return get('/news/view', { id })
  }
}

export const passportApi = {
  login(data) {
    return post('/passport/login', data)
  },
  register(data) {
    return post('/passport/register', data)
  },
  getMyDetail(params) {
    return get('/passport/my_detail', params)
  },
  editBase(data) {
    return post('/passport/edit_base', data)
  }
}

export const adminApi = {
  login(data) {
    return post('/admin/login', data)
  },
  userList(params) {
    return get('/admin/user_list', params)
  },
  enrollList(params) {
    return get('/admin/enroll_list', params)
  }
}
