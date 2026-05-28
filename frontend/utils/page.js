export const pageHelper = {
  dataset(e, field) {
    if (!e || !e.currentTarget || !e.currentTarget.dataset) return null
    return e.currentTarget.dataset[field]
  },

  url(e, ctx) {
    const url = this.dataset(e, 'url')
    if (url) {
      uni.navigateTo({ url })
    }
  },

  formatTime(date) {
    const year = date.getFullYear()
    const month = date.getMonth() + 1
    const day = date.getDate()
    const hour = date.getHours()
    const minute = date.getMinutes()
    const second = date.getSeconds()

    return `${[year, month, day].map(formatNumber).join('-')} ${[hour, minute, second].map(formatNumber).join(':')}`
  },

  formatDate(date, sep = '-') {
    const d = new Date(date)
    const year = d.getFullYear()
    const month = d.getMonth() + 1
    const day = d.getDate()
    return [year, month, day].map(formatNumber).join(sep)
  }
}

function formatNumber(n) {
  n = n.toString()
  return n[1] ? n : '0' + n
}