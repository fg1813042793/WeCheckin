# HBuilderX Android 调试指南

## 前置准备

### 1. 安装 HBuilderX
- 下载地址: https://www.dcloud.io/hbuilderx.html
- 推荐版本: HBuilderX 3.8+ (标准版或App开发版)

### 2. 安装必要插件
打开 HBuilderX → 工具 → 插件安装，安装以下插件:
- **uni-app 编译器** (必须)
- **App真机运行** (Android调试必须)

### 3. Android 设备准备
- **方式一: 真机调试**
  - Android 手机开启 **开发者选项** 和 **USB调试**
  - USB 连接电脑
  
- **方式二: 模拟器**
  - 安装 Android Studio 自带模拟器
  - 或安装 雷电模拟器/夜神模拟器 等

## 调试步骤

### 步骤1: 打开项目
1. 打开 HBuilderX
2. 文件 → 打开目录 → 选择 `WeCheckin` 项目文件夹
3. 等待项目初始化完成（右下角会显示"就绪"）

### 步骤2: 运行到 Android 设备
**方法A: 使用菜单栏**
```
运行 → 运行到手机或模拟器 → 运行到 Android App 基座
```

**方法B: 使用工具栏**
1. 点击工具栏的 **运行** 按钮 (▶️ 图标)
2. 选择 **运行到手机或模拟器**
3. 选择你的 Android 设备

**方法C: 快捷键**
- `Ctrl + R` (Windows/Linux)
- `Cmd + R` (Mac)

### 步骤3: 查看设备连接状态
- 成功连接后，HBuilderX 底部控制台会显示:
  ```
  [LOG] 正在连接设备...
  [LOG] 设备已连接: XXXXXXXX (Android XX)
  [LOG] 开始编译...
  ```

### 步骤4: 调试应用
- 应用会自动安装并启动到 Android 设备
- 可以使用 Chrome DevTools 进行调试:
  ```
  chrome://inspect
  ```
  或者在 HBuilderX 内置控制台查看日志

## 常见问题解决

### 问题1: 设备未识别
**解决方案:**
```bash
# Mac/Linux
adb devices

# Windows
adb.exe devices
```
确保输出包含你的设备ID。

### 问题2: USB调试未开启
**Android 设置路径:**
```
设置 → 关于手机 → 连续点击"版本号"7次
→ 返回设置 → 开发者选项 → 开启 USB调试
```

### 问题3: 编译错误
**检查清单:**
- [ ] Node.js 版本 >= 16
- [ ] npm install 已执行
- [ ] manifest.json 中 appid 已配置
- [ ] pages.json 路由配置正确

### 问题4: 白屏/闪退
**排查步骤:**
1. 查看 HBuilderX 控制台错误信息
2. 检查 manifest.json 权限配置
3. 确保 vue 组件语法正确

## 高级调试技巧

### 1. 实时热重载
修改代码后自动刷新，无需重新打包:
- HBuilderX 会自动检测文件变化
- 或手动按 `Cmd + R` 刷新

### 2. 断点调试
在 `.vue` 文件中设置断点:
```javascript
// 在需要调试的行号左侧点击
onLoad() {
  console.log('页面加载') // ← 点击这里设置断点
}
```

### 3. 查看网络请求
在 HBuilderX 控制台查看 API 请求:
```
[Network] GET /api/home/list 200 OK
[Network] POST /api/user/login 401 Unauthorized
```

### 4. 性能分析
使用 Chrome DevTools Performance 面板:
1. 打开 `chrome://inspect`
2. 选择对应的 WebView
3. 切换到 Performance 标签

## 生产环境打包

开发完成后，打包正式 APK:

### 方式一: HBuilderX 云打包
```
发行 → 原生App-云打包 → 选择 Android
```

### 方式二: 本地打包
```
发行 → 原生App-本地打包 → 生成本地打包资源
```
然后使用 Android Studio 打包 APK。

## 注意事项

1. **首次运行较慢**: 需要下载基座和编译资源
2. **保持USB连接**: 调试期间不要拔掉数据线
3. **权限申请**: 首次运行会请求相机、位置等权限
4. **日志过滤**: 控制台可使用过滤器只看关键信息

## 下一步

调试成功后可以:
- [ ] 测试所有页面功能
- [ ] 检查 UI 兼容性
- [ ] 测试不同屏幕尺寸
- [ ] 准备发布到应用商店