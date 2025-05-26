# Changelog

All notable changes to this project will be documented in this file.

## Types of Changes

- `Added` for new features
- `Changed` for changes in existing functionality
- `Deprecated` for soon-to-be removed features
- `Removed` for now removed features
- `Fixed` for any bug fixes
- `Security` in case of vulnerabilities
- `Dependency` for dependency updates
- `Performance` for performance improvements
- `Experimental` for experimental features

## [Unreleased]

### Added

- `Dockerfile` to containerize the application.

## [1.1.0] - 24.05.2025

### Added

- `sample` folder to save uploaded files.
- `.env` to store sensitive information and environment-specific configurations.
- `config/init.go` file which initializes env specific configurations.
- `helper/responseBuilder.go` for structuring the response.
- `version/version.go` to set the current version of project.
- `controllers/systemParametersController.go` in which I added constructors for getting `status` and `version`. Also added their endpoints in `routes/route.go`.

### Changed

* Renamed `controllers/controller.go` and `services/service.go` to `controllers/TFIDFController.go` and `services/TFIDFService.go` for better maintenance.

### Dependency

- Added `github.com/joho/godotenv` `v1.5.1` to manage env variables from .env file.
- Upgraded `GO` from `v1.24.0` to `v1.24.3`
