import { get, post } from '../utils/request'

export const homeApi = {
  getList(params) {
    return get('/home/list', params)
  },
  setupGet(params) {
    return get('/home/setup_get', params)
  }
}

export const enrollApi = {
  getList(params) {
    return get('/enroll/list', params)
  },
  detail(params) {
    return get('/enroll/view', params)
  },
  join(data) {
    return post('/enroll/join', data)
  },
  enrollSubmit(data) {
    return post('/enroll/enroll_submit', data)
  },
  joinDay(params) {
    return get('/enroll/join_day', params)
  },
  myJoinList(params) {
    return get('/enroll/my_join_list', params)
  },
  myUserList(params) {
    return get('/enroll/my_user_list', params)
  },
  myRecords(params) {
    return get('/enroll/my_records', params)
  },
  myCalendar(params) {
    return get('/enroll/my_calendar', params)
  },
  myDayRecords(params) {
    return get('/enroll/my_day_records', params)
  }
}

export const newsApi = {
  getList(params) {
    return get('/news/list', params)
  },
  detail(id) {
    return get('/news/view', { id })
  },
  cateList() {
    return get('/news/cate_list')
  }
}

export const geoApi = {
  reverse(params) {
    return get('/geo/reverse', params)
  }
}

export const passportApi = {
  login(data) {
    return post('/passport/login', data)
  },
  loginByPwd(data) {
    return post('/passport/login_pwd', data)
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

export const favApi = {
  list(params) {
    return get('/fav/my_list', params)
  },
  insert(data) {
    return post('/fav/update', data)
  },
  del(data) {
    return post('/fav/del', data)
  }
}

export const eventApi = {
  getList(params) {
    return get('/event/list', params)
  },
  getDetail(params) {
    return get('/event/view', params)
  },
  detail(params) {
    return get('/event/view', params)
  },
  participate(data) {
    return post('/event/participate', data)
  },
  myParticipate(params) {
    return get('/event/my_list', params)
  },
  myList(params) {
    return get('/event/my_list', params)
  },
  myRoles(params) {
    return get('/event/my_roles', params)
  },
  myManage(params) {
    return get('/event/my_managed', params)
  },
  myManaged(params) {
    return get('/event/my_managed', params)
  },
  dynamicList(params) {
    return get('/event/dynamics', params)
  },
  dynamics(params) {
    return get('/event/dynamics', params)
  },
  dynamicInsert(data) {
    return post('/event/dynamic_post', data)
  },
  dynamicPost(data) {
    return post('/event/dynamic_post', data)
  },
  scoreList(params) {
    return get('/event/scores', params)
  },
  scores(params) {
    return get('/event/scores', params)
  },
  scoreSave(data) {
    return post('/event/score_save', data)
  },
  participantList(params) {
    return get('/event/participant_list', params)
  }
}

export function userFormFields() {
  return get('/user_form_fields')
}
