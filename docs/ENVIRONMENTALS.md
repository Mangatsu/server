## 📝 Mangatsu Server - Configuration
Usable options inside the **.env** or **docker-compose.yml**:

_~~Struck out~~ values have no effect yet._

- ~~**MTSU_LOG_LEVEL**~~=info
  - Log level: error, warn, info, debug, trace
- **MTSU_INITIAL_ADMIN_NAME**=admin
- **MTSU_INITIAL_ADMIN_PW**=admin321
    - Credentials for the initial admin user. Recommended to change.
- **MTSU_HOSTNAME**=localhost
- **MTSU_PORT**=5050
    - Hostname and port for the server. Use **mtsuserver** as the hostname if using Docker Compose.
- **MTSU_BASE_PATHS**=freeform1;/home/user/doujinshi;;structured2:/home/user/manga
    - Paths to the archive directories. Relative or absolute paths are accepted.
    - First specify the type of the directory and a numerical ID (e.g. freeform1 or structured2) and then the path separated by a semicolon: `;`.
    - Multiple paths can be separated by a double-semicolon: `;;`.
    - Format: `<freeform|structured><ID>;<INTERNAL_PATH>;;<freeform|structured><ID>;<INTERNAL_PATH>`...
    - If using Docker Compose, make sure that the paths match the containerpaths in the volumes section.
- **MTSU_DATA_PATH**=../data
    - Location of the data dir which includes the SQLite db and the cache for gallery images and thumbnails. Relative or absolute paths are accepted.
    - Doesn't need changing if using Docker Compose.
- **MTSU_DISABLE_CACHE_SERVER**=false
  - True to disable the internal cache server (serves media files and thumbnails). Useful if one wants to use the web server such as NGINX to serve the files.
- ~~**MTSU_CACHE_SIZE**~~=10000
  - Max size of the cache where galleries are extracted from the library in MB. Can overflow a bit especially if set too low.
- ~~**MTSU_CACHE_TTL**~~=604800
  - Time to live for the cache in seconds (604800 = 1 week).
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

## 📝 Mangatsu Web - Configuration

- **NEXT_PUBLIC_MANGATSU_API_URL**=https://mangatsu-api.example.com
  - URL to the backend API server.
- **NEXT_MANGATSU_IMAGE_HOSTNAME**=mangatsu-api.example.com
  - Hostname or the domain where images are hosted. Usually the same as the domain in the API URL above.
- **NEXTAUTH_URL**=https://mangatsu.example.com
  - URL to the web client.
- **SECRET**=zb4DyuqELIfi8X0XRMa52yV9y6d011YTyyHBtczhYeHbzsquYnsr7Q9OkHNoLd6HvFqz4vvrCPcINol7sIEfLG6pY4D0KSHo
  - A random string used to hash tokens, sign cookies and generate cryptographic keys. Recommended to change. Can be generated with `openssl rand -base64 32`.
- **JWT_SIGNING_PRIVATE_KEY**='{"kty":"oct","kid":"bbdunoB1J71jPATtgJAseZnx36LB4ant1OfH6ysV78M","alg":"HS512","k":"axdN1SbiVxMIPJYapyHOpGuo7KcE_mkT6_bugy5xxG8"}'
  - Secret to sign JWTs in the web client. Recommended to change.
- **PORT**=3030
  - Port to run the web client on. If you change this and also use Docker Compose, remember to update the first port of frontend's ports in the `docker-compose.yml` file.
