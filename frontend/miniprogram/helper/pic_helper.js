/**
 * Notes: 图片处理相关函数
 * Ver : CCMiniCloud Framework 2.0.13 ALL RIGHTS RESERVED BY cclinux0730 (wechat)
 * Date: 2021-10-25 04:00:00 
 */


function getWritePhotosAlbum(callback) {
	wx.getSetting({
		success: res => {
			if (res.authSetting['scope.writePhotosAlbum']) {
				callback && callback();
			} else if (res.authSetting['scope.writePhotosAlbum'] === undefined) {
				wx.showModal({
					title: '提示',
					content: '您未开启保存图片到相册的权限，请点击确定去开启权限！',
					success: (res) => {
						if (res.confirm) {
							wx.authorize({
								scope: 'scope.writePhotosAlbum',
								success: (res) => {
									callback && callback()
								},
								fail: (res) => {
									wx.showToast({
										title: '您没有授权，无法保存到相册',
										icon: 'none'
									});
								}
							});
						} else {
						}
					}
				});
			} else {
				wx.showModal({
					title: '提示',
					content: '您未开启保存图片到相册的权限，请点击确定去开启权限！',
					success: (res) => {
						if (res.confirm) {
							wx.openSetting({
								success: (res) => {
									wx.showToast({
										icon: 'none',
										title: '正在保存图片',
									});
									if (res.authSetting['scope.writePhotosAlbum']) {
										callback && callback();
									} else {
										wx.showToast({
											title: '您没有授权，无法保存到相册！',
											icon: 'none'
										});
									}
								},
								fail: (res) => {
								}
							});
						} else {
							wx.showToast({
								title: '您没有授权，无法保存到相册',
								icon: 'none'
							});
						}
					}
				});
			}
		}
	});
}

module.exports = {
	getWritePhotosAlbum
}