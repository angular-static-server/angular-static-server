# Changelog

## [16.0.1](https://github.com/angular-static-server/angular-static-server/compare/16.0.0...16.0.1) (2023-08-27)


### Bug Fixes

* address timing issue with .env file watcher ([#16](https://github.com/angular-static-server/angular-static-server/issues/16)) ([6a2480e](https://github.com/angular-static-server/angular-static-server/commit/6a2480e3e02e4d7a0c483cf154d25ba7b2554810))
* **deps:** update angular monorepo to v16.2.1 ([#15](https://github.com/angular-static-server/angular-static-server/issues/15)) ([0fba0a0](https://github.com/angular-static-server/angular-static-server/commit/0fba0a0c0e3e675a3d7f0e3a07cab60f2a156018))
* **deps:** update angular monorepo to v16.2.2 ([#17](https://github.com/angular-static-server/angular-static-server/issues/17)) ([74ff2ad](https://github.com/angular-static-server/angular-static-server/commit/74ff2ad8171e2194f934259775adb9b17f852b5c))
* **deps:** update angularmaterial monorepo to v16.2.1 ([#11](https://github.com/angular-static-server/angular-static-server/issues/11)) ([b215def](https://github.com/angular-static-server/angular-static-server/commit/b215def2732dcd4deb79f918132366de952f252b))
* **deps:** update dependency tslib to v2.6.2 ([#12](https://github.com/angular-static-server/angular-static-server/issues/12)) ([02d0cef](https://github.com/angular-static-server/angular-static-server/commit/02d0cef20f5c35d358b1a8be4e1c31f6b7c47e33))

## 16.0.0 (2023-08-20)


### Features

* add X-Frame-Options header and configuration ([4f0f942](https://github.com/angular-static-server/angular-static-server/commit/4f0f9429f95250e10ff1617ef3a27439d8a05cda))
* configure image to be non-root and non-privileged ([83909f7](https://github.com/angular-static-server/angular-static-server/commit/83909f777afa3584a7dcf800530a1c0bfe19e823))
* extend CSP options ([358c555](https://github.com/angular-static-server/angular-static-server/commit/358c55578204eaf3a7185342d012fb733de856a5))
* implement Angular static server ([95678d7](https://github.com/angular-static-server/angular-static-server/commit/95678d7ec4986d03678450860fb2556aad8a4c10))
* implement Mozilla Dockerflow guidelines ([a00adde](https://github.com/angular-static-server/angular-static-server/commit/a00addef9d697cc8ddc22319f13d72046eac3fdd))


### Bug Fixes

* **deps:** update golang.org/x/exp digest to d852ddb ([#3](https://github.com/angular-static-server/angular-static-server/issues/3)) ([47de1a8](https://github.com/angular-static-server/angular-static-server/commit/47de1a88b43cce30e9ff8eb3aa117d120fada8e5))
* **deps:** update module github.com/urfave/cli/v2 to v2.25.7 ([#4](https://github.com/angular-static-server/angular-static-server/issues/4)) ([b5bbf2f](https://github.com/angular-static-server/angular-static-server/commit/b5bbf2f1b0ad5efa57b21ece4c797bdbb4f88d91))
* ensure nonce randomness is uniform ([5e623e4](https://github.com/angular-static-server/angular-static-server/commit/5e623e4f84a23293f8a0859ed57963eac98169c2))
* extended nonce to 16 characters, as recommended ([d35fcac](https://github.com/angular-static-server/angular-static-server/commit/d35fcac5e7f18bc466e6d65359e57ebc3e32ecce))
* relax Cache-Control setting from no-store to no-cache ([f4de8b0](https://github.com/angular-static-server/angular-static-server/commit/f4de8b04f949cd55490bcc8a872ad5d408978741))
* remove obsolete user/group creation in Dockerfile ([934a007](https://github.com/angular-static-server/angular-static-server/commit/934a007d6c9016d8d890b0d12f4c0acfdcb6d5f8))
