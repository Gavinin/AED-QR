# AED-QR (Automated External Defibrillator Quick Response)

<p align="center">
  <b>AED 二维码紧急救援系统</b><br>
  <i>为生命争取每一秒</i>
</p>

## 📖 简介 / Introduction

这是一个车载 AED（自动体外除颤器）的紧急二维码项目。

你可以将生成的二维码打印并贴在车身上。在紧急情况下，当有人需要使用 AED 进行抢救时，任何人都可以通过扫描该二维码，快速打开汽车的**前备箱、后备箱以及车门**，第一时间获取车内的 AED 设备，为挽救生命争取宝贵的时间。

### 🚗 支持的车辆
目前**仅支持特斯拉 (Tesla)**（因为作者自己开的是特斯拉 😂）。
非常欢迎其他汽车品牌的开发者参与贡献，进行适配！

### 🌍 支持的语言 (Supported Languages)
扫描后的界面支持多国语言，方便不同国家/地区的用户：
- 中文 (Chinese)
- 英文 (English)
- 法语 (French)
- 日语 (Japanese)
- 韩语 (Korean)
- 阿拉伯语 (Arabic)
- 西班牙语 (Spanish)

---

## 🚀 快速开始 / Quick Start

我们提供了开箱即用的 Docker 部署方式，非常简单：

### 1. 下载与安装
1. **下载最新版本**  
   前往项目的 [Releases 页面](https://github.com/Gavinin/AED-QR/releases)，下载最新发布的 `aed-qr-docker.zip` 压缩包。

2. **解压并进入目录**
   ```bash
   unzip aed-qr-docker.zip
   cd docker
   ```

3. **初始化环境**  
   运行提供的 setup 脚本：
   ```bash
   bash setup.sh
   ```

### 2. 配置参数 (关键步骤)
在启动服务前，你**必须**修改 `config/config.yml` 文件。以下是需要重点修改的字段：

- **`server.domain`**:  
  这是生成二维码时使用的域名地址。  
  *示例*: `http://your-public-ip:8080` 或 `https://your-domain.com`。  
  > ⚠️ **注意**: 请确保该地址在公网可访问，否则救援人员扫码后无法打开网页。

- **`jwt.secret`**:  
  请修改为一个随机字符串（至少 28 位），以确保安全。

- **`admin.username` & `admin.password`**:  
  强烈建议修改默认的用户名和密码，防止他人未经授权访问后台。

### 3. 启动服务
使用 Docker Compose 启动：
```bash
docker compose up -d
```

启动完成后，系统即可开始运行。

---

## 📖 使用指南

1. **进入后台管理**  
   打开浏览器访问你部署的地址（如 `http://your-ip:8080/login`），使用你在 `config.yml` 中配置的用户名和密码登录。

2. **添加车辆**  
   点击右下角的 **添加车辆** 按钮，根据提示完成授权（需要特斯拉账户）。

3. **获取二维码**  
   添加成功后，在车辆卡片的右侧会有一个 **二维码图标**，点击即可查看并下载你的紧急救援二维码。

4. **测试**  
   使用手机扫描二维码（模拟救援人员），确保能正常加载页面，并能成功点击打开车门/前备箱/后备箱。
   > ⚠️ **测试要求**: 你的手机必须能够访问到部署的服务器地址（公网访问）。

---

## 🛠 参与贡献 / Contributing

我们非常欢迎来自开源社区的贡献！如果你是一名开发者并希望让这个项目支持更多品牌的汽车，请提交 Pull Request！

---

## 📄 许可证 / License

本项目遵循 [MIT License](LICENSE) 许可证。