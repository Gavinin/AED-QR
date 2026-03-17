# AED-QR (Automated External Defibrillator Quick Response)

[🇨🇳 中文文档 (Chinese README)](docs/README_CN.md)

<p align="center">
  <b>AED QR Code Emergency Rescue System</b><br>
  <i>Fighting for every second to save a life</i>
</p>

## 📖 Introduction

This is an emergency QR code project for vehicle-mounted AEDs (Automated External Defibrillators).

You can print the generated QR code and attach it to your vehicle. In an emergency situation where someone needs an AED for rescue, anyone can scan this QR code to quickly open the vehicle's **frunk, trunk, and doors**, obtaining the AED equipment from inside the vehicle at the very first moment, fighting for precious time to save a life.

### 🚗 Supported Vehicles
Currently, **only Tesla is supported** (because the author drives a Tesla 😂).
Developers from other car brands are highly welcome to contribute and adapt the system!

### 🌍 Supported Languages
The scanned interface supports multiple languages to accommodate users from different countries/regions:
- English
- Chinese (中文)
- French (Français)
- Japanese (日本語)
- Korean (한국어)
- Arabic (العربية)
- Spanish (Español)

---

## 🚀 Quick Start

We provide an out-of-the-box Docker deployment method, which is very simple:

### 1. Download and Install
1. **Download the latest version**  
   Go to the project's [Releases page](https://github.com/Gavinin/AED-QR/releases) and download the latest released `aed-qr-docker.zip` archive.

2. **Unzip and enter the directory**
   ```bash
   unzip aed-qr-docker.zip
   cd docker
   ```

3. **Initialize the environment**  
   Run the provided setup script:
   ```bash
   bash setup.sh
   ```

### 2. Configuration (Crucial Step)
Before starting the service, you **must** edit the `config/config.yml` file. Here are the key fields you need to modify:

- **`server.domain`**:  
  This is the domain used for generating the QR code.  
  *Example*: `http://your-public-ip:8080` or `https://your-domain.com`.  
  > ⚠️ **Note**: Ensure this address is accessible from the public internet so that rescuers can open the link when scanning the code.

- **`jwt.secret`**:  
  Please change this to a random string (at least 28 characters) to ensure security.

- **`admin.username` & `admin.password`**:  
  It is strongly recommended to change the default username and password to prevent unauthorized access.

### 3. Start Service
Start it using Docker Compose:
```bash
docker compose up -d
```

Once started, the system will be up and running.

---

## 📖 Usage Guide

1. **Access the Admin Panel**  
   Open your browser and visit your deployed address (e.g., `http://your-ip:8080/login`). Log in using the username and password you configured in `config.yml`.

2. **Add Your Vehicle**  
   Click the **Add Vehicle** button in the bottom right corner and follow the prompts to complete the authorization (Tesla account required).

3. **Get the QR Code**  
   After successfully adding the vehicle, you will see a **QR Code** icon on the right side of the vehicle card. Click it to view and download your emergency QR code.

4. **Test**  
   Scan the QR code with your phone (simulate a rescuer). Ensure that the page loads correctly and that you can successfully unlock the vehicle doors/frunk/trunk.
   > ⚠️ **Test Requirement**: Your phone must be able to access the deployed server address (public network access).

---

## 🛠 Contributing

We highly welcome contributions from the open-source community! If you are a developer and want to make this project support more car brands, please submit a Pull Request!

---

## 📄 License

This project is licensed under the [MIT License](LICENSE).