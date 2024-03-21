## 📝 Mangatsu Server - Configuration
Usable options inside the **.env** or **docker-compose.yml**:

_~~Struck out~~ values have no effect yet._

- **MTSU_ENV**=production
    - Environment: production, development
- **MTSU_LOG_LEVEL**=info
    - Log level: debug, info, warn, error
- **MTSU_INITIAL_ADMIN_NAME**=admin
- **MTSU_INITIAL_ADMIN_PW**=admin321
    - Credentials for the initial admin user. Recommended to change.
- **MTSU_DOMAIN**: example.org.
    - For example, if the address for the server is "api.example.com", and for the frontend "read.example.com", the value here should be "example.com" for the cookies to work properly between subdomains.
    - https://developer.mozilla.org/en-US/docs/Web/HTTP/Cookies#define_where_cookies_are_sent
- **MTSU_STRICT_ACAO**: 'true'
    - When true, the server only allows authenticated connections from the MTSU_DOMAIN and its subdomains (eg .*.example.org). When false, connections from every origin is allowed.
    - https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Origin
- **MTSU_HOSTNAME**=localhost
- **MTSU_PORT**=5050
    - Hostname and port for the server. Use **mtsuserver** as the hostname if using Docker Compose.
- **MTSU_BASE_PATHS**=freeform1;/home/user/doujinshi;;structured2;/home/user/manga
    - Paths to the archive directories. Relative or absolute paths are accepted.
    - First specify the type of the directory and a numerical ID (e.g. freeform1 or structured2) and then the path separated by a semicolon: `;`.
    - Multiple paths can be separated by a double-semicolon: `;;`.
    - Format: `<freeform|structured><ID>;<INTERNAL_PATH>;;<freeform|structured><ID>;<INTERNAL_PATH>`...
    - If using Docker Compose, make sure that the paths match the containerpaths in the volumes section.
- **MTSU_DATA_PATH**=../data
    - Location of the data dir which includes the SQLite db and the cache for gallery images and thumbnails. Relative or absolute paths are accepted.
    - Doesn't need changing if using Docker Compose.
- **MTSU_DISABLE_CACHE_SERVER**=false
    - Set true to disable the internal cache server (serves media files and thumbnails). Useful if one wants to use the web server such as NGINX to serve the files.
- **MTSU_CACHE_TTL**=336h
    - Cache time to live (for example `336h` (2 weeks), `8h30m`). If a gallery is not viewed for this time, it will be purged from the cache.
- ~~**MTSU_CACHE_SIZE**~~=10000
    - Max size of the cache where galleries are extracted from the library in MB. Can overflow a bit especially if set too low.
- **MTSU_DB_NAME**=mangatsu
    - Name of the SQLite database file
- ~~**MTSU_DB**~~=sqlite
    - Database type: `sqlite`, `postgres`, `mysql` or `mariadb`
- ~~**MTSU_DB_HOST**~~=localhost
    - Hostname of the database server.
- ~~**MTSU_DB_PORT**~~=5432
    - Usually 5432 for PostgreSQL and 3306 for MySQL and MariaDB.
- ~~**MTSU_DB_USER**~~=mtsu-user
- ~~**MTSU_DB_PASSWORD**~~=s3cr3t
- **MTSU_VISIBILITY**=public
    - **public**: anyone can access the collection and its galleries.
    - **restricted**: users need a global passphrase to access collection and its galleries.
    - **private**: only logged-in users can access the collection and its galleries.
    - In all modes, user accounts are supported and have more privileges than anonymous users (e.g. favorite galleries).
- **MTSU_RESTRICTED_PASSPHRASE**=secretpassword
    - Passphrase to access the collection and its galleries.
    - Only used when **VISIBILITY** is set to **restricted**.
- **MTSU_REGISTRATIONS**=false
    - Whether to allow user registrations. If set to false, only admins can create new users.
    - **Currently, only affects the API path /register. Has no effect in the frontend.**
- **MTSU_JWT_SECRET**=secret123
    - Secret to sign JWTs for login sessions in the backend. Recommended to change.
- **MTSU_THUMBNAIL_FORMAT**=webp
  - Supported formats: webp 
  - AVIF support is planned. AVIF is said to take 20% longer to encode, but it compresses to 20% smaller size compared to WebP.

## 📝 Mangatsu Web - Configuration

- **NEXT_PUBLIC_MANGATSU_API_URL**=https://mangatsu-api.example.com
    - URL to the backend API server.
- **NEXT_PUBLIC_INTERNAL_MANGATSU_API_URL**=http://mtsuserver:5050
    - Internal URL to the backend server. Required when running both containers locally without external network.
    - For example, if both containers are running on the same network, the value should probably be "http://mtsuserver:5050".
- **NEXT_MANGATSU_IMAGE_HOSTNAME**=mangatsu-api.example.com
    - Hostname or the domain where images are hosted. Usually the same as the domain in the API URL.
- **PORT**=3030
    - Port to run the web client on. If you change this and also use Docker Compose, remember to update the first port of frontend's ports in the `docker-compose.yml` file.
