
<h1 align="center"> Mangatsu</h1>

<p align="center">
  <img src="docs/logo-small.png" />
</p>

## Still at very experimental stage. Expect lots of breaking changes.

> 🌕 Server application for storing doujinshi, manga, art collections and other galleries with API and user control. Written in Go.

### **[📰 CHANGELOG](docs/CHANGELOG.md)** | **[❤ CONTRIBUTING](docs/CONTRIBUTING.md)** | **[🎯 TODO](docs/TODO.md)**

## 📌 Installation and usage

### [📝 Configuration with environmentals](docs/ENVIRONMENTALS.md)
### [📚 Library directory structure](docs/LIBRARY.md)

### 🐳 Docker setup for both the server and web client (recommended)
- Set up a webserver of your choice. NGINX is recommended.
  - [Example config](docs/nginx.conf). The same config can be used for both the server and the web client. Just change the domains, SSL cert paths and ports.
- Install [Docker](https://docs.docker.com/engine/install/) (Linux, Windows or MacOS)
- Local archives
  - Download the [docker-compose.example.yml](docs/docker-compose.example.yml) and rename it to docker-compose.yml
  - Edit the docker-compose.yml file to your needs
  - Create data and archive directories
- Network archives with [Rclone](https://rclone.org)
  - Follow the [guide on Rclone site](https://rclone.org/docker/)
  - Download the [docker-compose.rclone.yml](docs/docker-compose.rclone.yml) and rename it to docker-compose.yml
- Run `docker-compose up` to start the server and web client

### 💻 Local setup
- Copy example.env as .env and change the values according to your needs.
- Create data and archive directories
- Build `go build`
- Run `backend` (`backend.exe` on Windows)

## 📌 Clients

### 🌏 Web client
- Included in the Docker setup above.
- Source: [Mangatsu Web](http://github.com/Mangatsu/web)

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
