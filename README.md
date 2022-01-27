
<h1 align="center"> Mangatsu</h1>

<p align="center">
  <img src="docs/logo-small.png" />
</p>

## Still at very experimental stage. Expect lots of breaking changes.

> 🌕 Server application for storing doujinshi, manga, art collections and other galleries with API and user control. Written in Go.

### **[📰 CHANGELOG](docs/CHANGELOG.md)** | **[❤ CONTRIBUTING](docs/CONTRIBUTING.md)** | **[🎯 TODO](docs/TODO.md)**

## 📌 Installation and usage

### 🐳 Docker setup (recommended)
- Coming soon for both the server and the web app

### 💻 Local setup
- Coming soon

## 📌 Clients

### 🖥 Web interface
- [Mangatsu Web](http://github.com/Mangatsu/web)

### 📱 [Tachiyomi](https://tachiyomi.org) extension for Android
- Coming soon

## 📌 Features
- Organizing and tagging local (and remote with tools like [rclone](https://rclone.org)) collections of manga, doujinshi and other art
  - **Mangatsu will never do any writes inside the archive location.**
  - Supports **ZIP** (or CBZ), **RAR** (or CBR) and plain image (png, jpg, jpeg, webp, gif, tiff, bmp) files. 
    - 7z, PDF and video support is planned.
- Metadata parsing from filenames, JSON/TXT files (inside or beside the archive). More to come.
- API-access to the collection and archives
  - Extensive filtering, sorting and searching capabilities.
  - Additional features for registered users such as tracking reading progress and adding favorite groups.
- User access control
  - **Private**: only logged-in users can access the collection and archives (**registration can be disabled**).
  - **Restricted**: users need a global passphrase to access collection and its galleries.
  - **Public**: anyone can access (only read) collection and its galleries.

- Users with roles (admin, member, viewer) and indefinite login sessions with the option to log out or delete them
- Local cache and thumbnail support

## 📌 Usage

### 📝 Configuration
Options inside the .env file:

- **MTSU_INITIAL_ADMIN_NAME**=admin
- **MTSU_INITIAL_ADMIN_PW**=admin321
  - Credentials for the initial admin user.
- **MTSU_HOSTNAME**=localhost
- **MTSU_PORT**=5050
  - Hostname and port for the server.
- **MTSU_BASE_PATHS**=freeform;/home/user/unstructured-manga;;structured:/home/user/structured-manga
  - Root paths to the collection and archives. Relative or absolute paths are accepted.
  - First specify the type of the directory (freeform or structured) and then the path separated by a semicolon: `;`
  - Multiple paths can be separated by a double-semicolon: `;;`
- **MTSU_DATA_PATH**=../data
  - Location of the data dir which includes the SQLite db and cache for gallery images and thumbnails. Relative or absolute paths are accepted.
- **MTSU_VISIBILITY**=public
  - **public**: anyone can access the collection and its galleries
  - **restricted**: users need a global passphrase to access collection and its galleries
  - **private**: only logged-in users can access the collection and its galleries
- **MTSU_RESTRICTED_PASSPHRASE**=secret321
  - Passphrase to access the collection and its galleries.
  - Only used when **VISIBILITY** is set to **restricted**.
- **MTSU_REGISTRATIONS**=false
  - Whether to allow user registrations. If set to false, only admins can create new users. **Currently, only affects the API path /register. Has no effect in the frontend.**
- **MTSU_JWT_SECRET**=secret123
  - Secret value for signing JWT tokens for login sessions in the backend.

### 📂 Directory structure
**Multiple root directories are supported.** I suggest creating a structured format for proper long-running manga, and a freeform structure for doujinshi and other art collections. Examples follow:

- **Freeform**: galleries can be up to three levels deep. Good for doujinshi, one-shots and other more unstructured collections.
  - External JSON metadata files have to be placed in the same level as the gallery archive. Preferably having the same name as the gallery archive. If no exact match is found, filename close enough will be used instead.
  - **Option to create categories according to the 1st level of the directory structure.** In the example, doujinshi and art would be created as categories. The last lonely archive would be uncategorized.

```
📂 freeform
├── 📂 doujinshi
│       ├──── 📂 oppai
│       │     ├──── 📦 [Group (Artist)] Ecchi Doujinshi.cbz
│       │     └──── 📄 [Group (Artist)] Ecchi Doujinshi.json
│       ├──── 📦 (C99) [Group (Artist)] elfs.zip
│       ├──── 📄 (C99) [Group (Artist)] elfs.json
│       └──── 📦 (C88) [kroup (author, another author)] Tankoubon [DL].zip  (JSON or TXT metafile inside)
├── 📂 art
│       ├──── 📂 [Artist] Pixiv collection
│       │     ├──── 🖼️ 0001.jpg
│       │     ├────...
│       │     └──── 🖼️ 0300.jpg
│       ├──── 📦 art collection y.rar
│       └──── 📄 art collection y.json
└── 📦 (C93) [group (artist)] Lonely doujinshi (Magical Girls).cbz
```

- **Structured**: galleries follow a strict structure. Good for long-running manga (shounen, seinen etc).
  - `Manga -> Volumes -> Chapters`, `Manga -> Volumes` or `Manga -> Chapters`
  - Galleries' Series is set to the name of the 1st level directory except for galleries in the root directory.

```
📂 structured
├── 📕 Manga 1
│       ├── 📦 Volume 1.cbz
│       ├── 📦 Volume 2.cbz
│       ├── 📦 Volume 3.cbz
│       └── 📦 Volume 4.zip
├── 📘 Manga 2
│       └── 📂 Vol. 1
│           ├── 📦 Chapter 1.zip
│           ├── 📦 Chapter 2.zip
│           └── 📦 Chapter 3.rar
├── 📗 Manga 3
│       ├── 📦 Chapter 1.zip
│       ├── ...
│       └── 📦 Chapter 30.rar
└── 📦 One Shot Manga.rar
```
