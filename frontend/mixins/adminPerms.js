export default {
  data() {
    return {
      perms: [],
      adminInfo: null
    }
  },

  onLoad() {
    this.loadPermsFromStorage()
  },

  onShow() {
    this.loadPermsFromStorage()
  },

  methods: {
    loadPermsFromStorage() {
      this.perms = uni.getStorageSync('admin_perms') || []
      this.adminInfo = uni.getStorageSync('admin_info') || null
    },

    hasPerm(perm) {
      if (this.adminInfo && this.adminInfo.type == 1) return true
      return this.perms.indexOf(perm) !== -1
    }
  }
}
