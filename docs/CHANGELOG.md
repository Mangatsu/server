# Changelog

All notable changes of this project will be documented in this file.

> The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres
> to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.7.2] - 2024-01-22

### Added

- Automatically migrating bcrypt passwords to argon2
- Username, password, session name, cookie age validations

### Changed

- Bcrypt to Argon2id for hashing passwords

## [0.7.1] - 2024-01-21

### Added

- Return expiredIn value (unix time) for anonymous login as well
- On logout, clear the cookie even if the token is incorrect
- Status gone (410) to errorHandler

### Fixed

- Regression where restricted/anonymous access was not working
- A bug with returning 401 when not needed
- Setting thumbnail count when generating thumbnails

### Changed

- Updated dependencies

## [0.7.0] - 2024-01-09

### Added

- **Endpoint to return task status `/status`**
- **Query param "meta" to read the metadata without having to read the gallery itself**
- Return current session ID on login
- `deleted` and `page_thumbnails` column to gallery table
- tif (tiff) and heif image extensions to supported extensions

### Fixed

- Error message casing
- Regression where the API would return an error when a gallery was not found
-

### Changed

- Minimum Go version to 1.21
- Updated dependencies
- Enabled Profile-guided optimization ([go.dev/doc/pgo](https://go.dev/doc/pgo))
- Regenerated jet models
- Deprecated rand.Seed to rand.New
- Deprecated ioutil.ReadDir to os.ReadDir

### Docs

- Instructions to install goose and jet tools

## [0.6.1] - 2023-05-30

**Released packages also on GHCR (GitHub Container Registry) alongside DockerHub:**

- [ghcr.io/mangatsu/server](https://github.com/Mangatsu/server/pkgs/container/server)
- [ghcr.io/mangatsu/web](https://github.com/Mangatsu/server/pkgs/container/server)

### Added

- More secure logic to handle CORS
    - An env `MTSU_STRICT_ACAO` ('true' or 'false') to disable or enable it

### Changed

- Update JWT package to v5

## [0.6.0] - 2023-05-29

### Added

- `MTSU_ENV` environmental: `development` or `production`
- `MTSU_LOG_LEVEL` environmental: `debug`, `info`, `warn` or `error`

### Changed

- Replace [logrus](https://github.com/sirupsen/logrus) with [zap](https://github.com/uber-go/zap)
    - and refactor accordingly

## [0.5.1] - 2023-05-28

### Added

- New environmental: MTSU_DOMAIN
- Domain property to Cookies send back to browser
    - As, if Domain is specified, then subdomains are always included in the allowed domains.
    - https://developer.mozilla.org/en-US/docs/Web/HTTP/Cookies#define_where_cookies_are_sent

### Fixed

- Internal version number

## [0.5.0] - 2023-05-28

### Added

- Support for cookies (JWT)
- Support for 7zip (7z) galleries

### Fixed

- Broken response when there were no galleries

### Changed

- Minimum Go version to 1.20
- Updated dependencies

### Docs

- Remove next-auth references
    - next-auth will be removed in Mangatsu Web v0.5.0

## [0.4.3] - 2022-04-17

### Added

- Periodically prune sessions

### Changed

- Increase maximum session duration to 1 year

## [0.4.2] - 2022-04-16

### Added

- Show internal server error messages in console
- API endpoint for gallery count `/galleries/count`
- Seed parameter for shuffling gallery results

### Fixed

- Setting a custom SQLite db filename
- Updating galleries internally
- Updating translated and native titles

### Changed

- Return only structured & "non-null Series" galleries when grouping

### Docs

- Update preview images

## [0.4.1] - 2022-04-15

### Fixed

- Regression of not being able to log in with passphrase

## [0.4.0] - 2022-04-15

### Added

- Periodically prune the gallery cache of old entries
    - Time to live can be configured via `MTSU_CACHE_TTL` environment variable. Defaults to 336h (14 days)
    - Utilizes mutex to prevent reading and deleting the same entry at the same time
- Environmental variable to disable migrations `MTSU_DB_MIGRATIONS=false`
- Environment variable to override the default database name: `MTSU_DB_NAME=mangatsu`
- Embedded migrations to binary

### Changed

- Updated dependencies

### Docs

- Cross out 7z in README as it's not supported yet
- Fix mistakes in example.env and ENVIRONMENTALS.md

## [0.3.1] - 2022-03-24

### Fixed

- Regression where order direction of galleries would not work
- Version number in API response

### Docs

- Add preview images to README

## [0.3.0] - 2022-03-20

### Added

- Support for returning galleries grouped by Series from the API
- Hath (H@H) and EHDL meta text file parsers
- Support for calling all metadata parsers through the API
- Validations when updating gallery

### Fixed

- GetTags SQL query

### Changed

- Updated dependencies
- Updated Go to 1.18
- Harden title language name
  parsing ([list of supported languages](https://github.com/Mangatsu/server/blob/main/pkg/metadata/language.go))
- Disallow empty namespaces or names in tags

## [0.2.0] - 2022-02-27

### Added

- UpdateGallery API method

### Fixed

- Error when a single (e.g. Gallery) result is empty
- Returning favorite groups could include empty groups

### Changed

- Renamed TitleShort column to TitleTranslated
- Updated dependencies

### Removed

- Artists and Circle columns from Gallery
    - Not needed as the same can be achieved by using the `artist` and `circle` namespace tags

## [0.1.2] - 2022-02-03

### Fixed

- Resetting favorite groups
- Returning favorite groups could include empty groups

## [0.1.1] - 2022-02-03

### Changed

- Change HTTP method of updating progress and favorites from PUT to PATCH

### Fixed

- Filtering by favorite groups

## [0.1.0] - 2022-01-31

### Added

- Initial release
