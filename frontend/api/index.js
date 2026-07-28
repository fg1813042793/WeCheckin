import { get, post, postJSON, put, del } from '../utils/request'

const API_V2 = '/api/v2'

function valueOf(source, keys = ['id'], fallback = '') {
  if (source === null || source === undefined) return fallback
  if (typeof source !== 'object') return source
  for (const key of keys) {
    const value = source[key]
    if (value !== undefined && value !== null && value !== '') return value
  }
  return fallback
}

function pathParam(source, keys = ['id'], fallback = '') {
  return encodeURIComponent(String(valueOf(source, keys, fallback)))
}

function setupParams(params = {}) {
  return params && params.params ? params.params : params
}

export const homeApi = {
  getList(params) {
    return get(`${API_V2}/home`, params)
  },
  setupGet(params) {
    return get(`${API_V2}/home/setup`, setupParams(params))
  }
}

export const enrollApi = {
  getList(params) {
    return get(`${API_V2}/enrollments`, params)
  },
  detail(params) {
    return get(`${API_V2}/enrollments/${pathParam(params)}`, params)
  },
  join(data) {
    return post(`${API_V2}/enrollments/${pathParam(data, ['enroll_id', 'enrollId', 'id'])}/joins`, data)
  },
  enrollSubmit(data) {
    return post(`${API_V2}/enrollments/${pathParam(data, ['enroll_id', 'enrollId', 'id'])}/submissions`, data)
  },
  joinDay(params) {
    return get(`${API_V2}/enrollments/${pathParam(params, ['id', 'enroll_id', 'enrollId'])}/join-days`, params)
  },
  myJoinList(params) {
    return get(`${API_V2}/me/enrollments`, params)
  },
  myUserList(params) {
    return get(`${API_V2}/me/enrollment-users`, params)
  },
  myRecords(params) {
    return get(`${API_V2}/me/enrollment-records`, params)
  },
  myCalendar(params) {
    return get(`${API_V2}/me/enrollment-calendar`, params)
  },
  myDayRecords(params) {
    return get(`${API_V2}/me/enrollment-day-records`, params)
  }
}

export const newsApi = {
  getList(params) {
    return get(`${API_V2}/news`, params)
  },
  detail(id) {
    return get(`${API_V2}/news/${pathParam(id)}`)
  },
  cateList() {
    return get(`${API_V2}/news/categories`)
  }
}

export const geoApi = {
  reverse(params) {
    return get(`${API_V2}/geo/reverse`, params)
  }
}

export const passportApi = {
  login(data) {
    return post(`${API_V2}/auth/login`, data)
  },
  loginByPwd(data) {
    return post(`${API_V2}/auth/password-login`, data)
  },
  register(data) {
    return post(`${API_V2}/auth/register`, data)
  },
  getMyDetail(params) {
    return get(`${API_V2}/me`, params)
  },
  editBase(data) {
    return put(`${API_V2}/me`, data)
  },
  getPhone(data) {
    return post(`${API_V2}/me/phone`, data)
  },
  logout(data) {
    return post(`${API_V2}/me/logout`, data)
  }
}

export const favApi = {
  list(params) {
    return get(`${API_V2}/me/favorites`, params)
  },
  insert(data) {
    return post(`${API_V2}/me/favorites`, data)
  },
  del(data) {
    return del(`${API_V2}/me/favorites/${pathParam(data, ['oid', 'id'])}`, data)
  },
  check(params) {
    return get(`${API_V2}/me/favorites/check`, params)
  }
}

export const eventApi = {
  getList(params) {
    return get(`${API_V2}/events`, params)
  },
  getDetail(params) {
    return get(`${API_V2}/events/${pathParam(params)}`, params)
  },
  detail(params) {
    return get(`${API_V2}/events/${pathParam(params)}`, params)
  },
  participate(data) {
    return post(`${API_V2}/events/${pathParam(data, ['event_id', 'eventId', 'id'])}/participants`, data)
  },
  myParticipate(params) {
    return get(`${API_V2}/me/events`, params)
  },
  myList(params) {
    return get(`${API_V2}/me/events`, params)
  },
  myRoles(params) {
    return get(`${API_V2}/me/event-roles`, params)
  },
  myManage(params) {
    return get(`${API_V2}/me/managed-events`, params)
  },
  myManaged(params) {
    return get(`${API_V2}/me/managed-events`, params)
  },
  dynamicList(params) {
    return get(`${API_V2}/events/${pathParam(params, ['event_id', 'eventId', 'id'])}/dynamics`, params)
  },
  dynamics(params) {
    return get(`${API_V2}/events/${pathParam(params, ['event_id', 'eventId', 'id'])}/dynamics`, params)
  },
  dynamicInsert(data) {
    return post(`${API_V2}/events/${pathParam(data, ['event_id', 'eventId', 'id'])}/dynamics`, data)
  },
  dynamicPost(data) {
    return post(`${API_V2}/events/${pathParam(data, ['event_id', 'eventId', 'id'])}/dynamics`, data)
  },
  scoreList(params) {
    return get(`${API_V2}/events/${pathParam(params, ['event_id', 'eventId', 'id'])}/scores`, params)
  },
  scores(params) {
    return get(`${API_V2}/events/${pathParam(params, ['event_id', 'eventId', 'id'])}/scores`, params)
  },
  scoreSave(data) {
    return post(`${API_V2}/events/${pathParam(data, ['event_id', 'eventId', 'id'])}/scores`, data)
  },
  participantList(params) {
    return get(`${API_V2}/events/${pathParam(params, ['event_id', 'eventId', 'id'])}/participants`, params)
  }
}

export function userFormFields() {
  return get(`${API_V2}/user-form-fields`)
}

export const formkitApi = {
  listTypes() {
    return get(`${API_V2}/admin/survey-types`)
  },
  parseSchema(data) {
    return post(`${API_V2}/admin/survey-schema/parse`, data)
  },
  eval(data) {
    return post(`${API_V2}/admin/survey-expressions/evaluate`, data)
  },
  validate(data) {
    return postJSON(`${API_V2}/survey/validate`, data)
  },
  apply(data) {
    return postJSON(`${API_V2}/survey/apply`, data)
  }
}

export const examApi = {
  getList(params) {
    return get(`${API_V2}/exams`, params)
  },
  getDetail(params) {
    if (params && params.session && !valueOf(params, ['id', 'examId'])) {
      return get(`${API_V2}/exam-results`, params)
    }
    return get(`${API_V2}/exams/${pathParam(params, ['id', 'examId'])}`, params)
  },
  start(data) {
    return post(`${API_V2}/exams/${pathParam(data, ['examId', 'id'])}/start`, data)
  },
  saveAnswer(data) {
    return put(`${API_V2}/exam-records/${pathParam(data, ['recordId', 'id'])}/answers`, data)
  },
  submit(data) {
    return postJSON(`${API_V2}/exams/${pathParam(data, ['examId', 'id'], 0)}/submissions`, data)
  },
  validate(data) {
    return postJSON(`${API_V2}/exams/${pathParam(data, ['examId', 'id'])}/validation`, data)
  },
  getRecord(params) {
    return get(`${API_V2}/exam-records/${pathParam(params, ['recordId', 'id'])}`, params)
  },
  myRecords() {
    return get(`${API_V2}/me/exam-records`)
  }
}

export const surveyApi = {
  getList(params) {
    return get(`${API_V2}/surveys`, params)
  },
  getDetail(params) {
    return get(`${API_V2}/surveys/${pathParam(params, ['surveyId', 'id'])}`, params)
  },
  apply(data) {
    return postJSON(`${API_V2}/survey/apply`, data)
  },
  validate(data) {
    return postJSON(`${API_V2}/survey/validate`, data)
  },
  submit(data) {
    return postJSON(`${API_V2}/surveys/${pathParam(data, ['surveyId', 'id'])}/responses`, data)
  },
  myResponses() {
    return get(`${API_V2}/me/survey-responses`)
  },
  myResponse(params) {
    return get(`${API_V2}/me/survey-responses/${pathParam(params)}`, params)
  }
}
