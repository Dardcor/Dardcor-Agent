# 🚀 Dardcor Agent

Dardcor Agent adalah **AI programming assistant otonom tingkat tinggi** yang dirancang untuk efisiensi token maksimal dan eksekusi sistem secara mandiri.

Agent ini dapat berjalan sebagai:

* ⚡ **Gateway Server**
* 💻 **Interactive CLI (Command Line Interface)**
* 🌐 **Web UI modern (real-time)**

---

## ✨ Highlights

* 🔥 **Ultra Token Efficient**
  Menghasilkan output logika maksimal dengan penggunaan token seminimal mungkin.

* 🤖 **Fully Autonomous Execution**
  Mampu menganalisis, memperbaiki, dan menjalankan operasi secara mandiri di sistem lokal.

* 🧩 **Unified Interface**
  Satu sistem, dua mode:

  * Web UI (visual & interaktif)
  * CLI (cepat & ringan)

* 🛠️ **Self-Healing System**
  Dilengkapi *system doctor* untuk mendeteksi dan memperbaiki error secara otomatis.

---

## 📦 Installation

### 🔧 Manual Installation (Recommended)

```bash
git clone https://github.com/dardcor/dardcor-agent.git
cd dardcor-agent
```

📌 **Catatan:**

* Instalasi global (`-g`) akan:

  * Menambahkan command `dardcor` ke terminal
  * Menginisialisasi database lokal berbasis JSON otomatis

---

### 🔧 Manual Installation (Global)

```bash
npm install -g dardcor-agent
```

### 🧹 Uninstall

```bash
npm uninstall -g dardcor-agent
```

---

## ⚡ Usage

Dardcor Agent memiliki command minimalis untuk menghindari konflik dan menjaga efisiensi sistem.

---

### 📖 Menampilkan Semua Command

```bash
dardcor help
```

---

### 🛠️ Self-Repair System

Jika terjadi error atau sistem tidak bisa berjalan:

```bash
dardcor doctor
```

🧠 Agent akan:

* Mendeteksi konfigurasi rusak
* Memperbaiki dependency
* Mengembalikan sistem ke kondisi optimal

---

### 🌐 Jalankan Web UI (Server Mode)

```bash
dardcor run
```

💡 Fitur:

* Dashboard real-time
* Cache-busting system
* UI interaktif

📍 Akses di:

```
http://127.0.0.1:25000
```

---

### 💻 Jalankan CLI Mode

```bash
dardcor cli
```

⚡ Mode ini cocok untuk:

* Coding cepat
* Automasi terminal
* Workflow ringan tanpa UI

---

## 🧠 System Architecture

* ⚙️ **Config File**
  Disimpan di:

  ```
  ~/.dardcor/config.json
  ```

* 🗄️ **Local Database**
  Berbasis file, berada di folder:

  ```
  /database
  ```

* 🔒 **Secure Local Network Only**

  * Semua aktivitas jaringan dibatasi ke:

  ```
  127.0.0.1:25000
  ```

  * Tidak ada akses eksternal → lebih aman & bebas konflik port

---

## 🎯 Design Philosophy

> **"Minimal input, maximum execution."**

Fokus utama:

* Efisiensi
* Otonomi
* Stabilitas sistem lokal

---

## ⚠️ Notes

* Pastikan Node.js sudah terinstall
* Gunakan terminal dengan akses admin jika diperlukan
* Direkomendasikan untuk developer yang ingin workflow otomatis & cepat

---

## 🧩 Future Vision (Optional)

* Integrasi multi-agent system
* Cloud sync (optional, tetap privacy-first)
* Plugin ecosystem

---

## 🧑‍💻 Author

Developed by **Dardcor** 🚀
