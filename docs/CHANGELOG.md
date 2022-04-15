# Changelog

All notable changes of this project will be documented in this file.

> The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.4.0] - 2022-04-15

### Added
- Periodically prune the gallery cache of old entries
  - Time to live can be configured via `MTSU_CACHE_TTL` environment variable. Defaults to 336h (14 days)
  - Utilizes mutex to prevent reading and deleting the same entry at the same time
- Environmental variable to disable migrations `MTSU_DB_MIGRATIONS=false`
- Environment variable to override the default database name: `MTSU_DB=mangatsu`
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
- Harden title language name parsing ([list of supported languages](https://github.com/Mangatsu/server/blob/main/pkg/metadata/language.go))
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
