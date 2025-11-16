# [1.0.0](https://github.com/royisme/BobaMixer/compare/v1.0.10...v1.0.0) (2025-11-16)


### Bug Fixes

* address golint errors in i18n implementation ([e6f9da5](https://github.com/royisme/BobaMixer/commit/e6f9da57a0c7c29648615c563dc7e8dd28260a2c))
* provide valid default profiles.yaml template on first initialization ([1e39d64](https://github.com/royisme/BobaMixer/commit/1e39d64636cf7ad8eae304f0fe31d3dc83b6c8fe))


### Features

* add adaptive theme system and i18n support following Bubble Tea best practices ([64688b0](https://github.com/royisme/BobaMixer/commit/64688b0f56f35f260d5c12fb5e43d74974b61ec1))
* add friendly welcome screen for TUI when profiles are missing ([c234408](https://github.com/royisme/BobaMixer/commit/c234408c556e53a77267f471ae66ab39992490db))
* integrate adaptive theme and i18n into TUI ([0fd193f](https://github.com/royisme/BobaMixer/commit/0fd193fc5209ab0815447f5ceac1d654d2241381))



# [1.0.0](https://github.com/royisme/BobaMixer/compare/v1.0.9...v1.0.0) (2025-11-16)



# [1.0.0](https://github.com/royisme/BobaMixer/compare/v1.0.8...v1.0.0) (2025-11-15)


### Bug Fixes

* correct parseRows implementation in suggestions store ([d895fd1](https://github.com/royisme/BobaMixer/commit/d895fd1aeeac90a410271f70f894a4b01e4df1c5))
* correct version progression in database migration ([e61a4e2](https://github.com/royisme/BobaMixer/commit/e61a4e2ddea784c46c6b53ba373acd6e9e8ba494))
* improve routing DSL and suggestion parsing ([3d43c3f](https://github.com/royisme/BobaMixer/commit/3d43c3fe1d167f9cbc52fa0ec820eb234278c744))
* prevent zero-priced models and dead links ([c898096](https://github.com/royisme/BobaMixer/commit/c898096a64a5b343fdfd6cbef0118b2885016934))
* remove unused daysAgo parameter from insertLatencySession ([1b18493](https://github.com/royisme/BobaMixer/commit/1b18493fb590196c05fd7ba3a3b8db45c25c9993))
* resolve all linter warnings for CI compliance ([db18f82](https://github.com/royisme/BobaMixer/commit/db18f82dd645ebcb243311e8251275804ae06059))
* resolve golint and typecheck errors ([479576f](https://github.com/royisme/BobaMixer/commit/479576f392ea83689431de5d2717de2138710216))
* resolve routing ctx_chars matching and linter issues ([aa06eb2](https://github.com/royisme/BobaMixer/commit/aa06eb24059f2f10715ccd672ff4f3103d3e9453))
* respect pricing.yaml sources configuration in Load() ([c1e26a2](https://github.com/royisme/BobaMixer/commit/c1e26a2743b93e25daa64c3c529497a76f15018b))


### Features

* add boba doctor --pricing command and fix linter issues ([4be12c9](https://github.com/royisme/BobaMixer/commit/4be12c9b44014132c9c86aa104f678f788f646c7))
* add foundational logging, secrets, and db helpers ([5ac9286](https://github.com/royisme/BobaMixer/commit/5ac92861d46e2764c2488b0f676f9e0786cb9478))
* add HTTP retry and enhanced diagnostics (batch 3) ([b0691a9](https://github.com/royisme/BobaMixer/commit/b0691a9eda3d0fe5fb56c6679b7fd5a0ef870027))
* align routing package with TDD spec (M3 partial) ([cc2ec37](https://github.com/royisme/BobaMixer/commit/cc2ec37261d8b8a49d71d482ab5c66c0d97429a6))
* enhance doctor diagnostics and tooling setup ([e852a97](https://github.com/royisme/BobaMixer/commit/e852a976b35e030391ecc9a9131719d5a4946ecd))
* implement comprehensive pricing.Load system with multi-source support ([fda8939](https://github.com/royisme/BobaMixer/commit/fda8939083c62ea10b80af567a9934bbd62e87a7))
* implement core execution features (batch 2) ([31d47fe](https://github.com/royisme/BobaMixer/commit/31d47fed6605a5cd52b95ca2e4a9f9c85b2b7018))
* implement DSL conditions, exploration mode, and feature flags ([2297a21](https://github.com/royisme/BobaMixer/commit/2297a2175d948e27ecee82dc0ad5d0f5cdc87936))
* implement httpx adapter and align stats APIs (M2-M3) ([85bfb4d](https://github.com/royisme/BobaMixer/commit/85bfb4dd8c7808f66404ca64254aeb29ed9b5752))
* implement structured logging and connect TUI dashboard ([5715034](https://github.com/royisme/BobaMixer/commit/5715034c8792521da69ed515590ed4f6f3dc4c85)), closes [#P0-4](https://github.com/royisme/BobaMixer/issues/P0-4) [#P2-5](https://github.com/royisme/BobaMixer/issues/P2-5) [#P1-2](https://github.com/royisme/BobaMixer/issues/P1-2) [#P3-3](https://github.com/royisme/BobaMixer/issues/P3-3)
* implement TDD spec core modules (M1-M2 partial) ([b71fea3](https://github.com/royisme/BobaMixer/commit/b71fea3dcf714c54b8277aa0919ccff4a93a641c))
* prioritize remote pricing before cache ([404ee5e](https://github.com/royisme/BobaMixer/commit/404ee5e72381a55ba1e1540d2df4be563ebb4af7))



# [1.0.0](https://github.com/royisme/BobaMixer/compare/v1.0.8...v1.0.0) (2025-11-15)


### Bug Fixes

* correct parseRows implementation in suggestions store ([d895fd1](https://github.com/royisme/BobaMixer/commit/d895fd1aeeac90a410271f70f894a4b01e4df1c5))
* correct version progression in database migration ([e61a4e2](https://github.com/royisme/BobaMixer/commit/e61a4e2ddea784c46c6b53ba373acd6e9e8ba494))
* improve routing DSL and suggestion parsing ([3d43c3f](https://github.com/royisme/BobaMixer/commit/3d43c3fe1d167f9cbc52fa0ec820eb234278c744))
* prevent zero-priced models and dead links ([c898096](https://github.com/royisme/BobaMixer/commit/c898096a64a5b343fdfd6cbef0118b2885016934))
* resolve all linter warnings for CI compliance ([db18f82](https://github.com/royisme/BobaMixer/commit/db18f82dd645ebcb243311e8251275804ae06059))
* resolve golint and typecheck errors ([479576f](https://github.com/royisme/BobaMixer/commit/479576f392ea83689431de5d2717de2138710216))
* resolve routing ctx_chars matching and linter issues ([aa06eb2](https://github.com/royisme/BobaMixer/commit/aa06eb24059f2f10715ccd672ff4f3103d3e9453))
* respect pricing.yaml sources configuration in Load() ([c1e26a2](https://github.com/royisme/BobaMixer/commit/c1e26a2743b93e25daa64c3c529497a76f15018b))


### Features

* add boba doctor --pricing command and fix linter issues ([4be12c9](https://github.com/royisme/BobaMixer/commit/4be12c9b44014132c9c86aa104f678f788f646c7))
* add foundational logging, secrets, and db helpers ([5ac9286](https://github.com/royisme/BobaMixer/commit/5ac92861d46e2764c2488b0f676f9e0786cb9478))
* add HTTP retry and enhanced diagnostics (batch 3) ([b0691a9](https://github.com/royisme/BobaMixer/commit/b0691a9eda3d0fe5fb56c6679b7fd5a0ef870027))
* align routing package with TDD spec (M3 partial) ([cc2ec37](https://github.com/royisme/BobaMixer/commit/cc2ec37261d8b8a49d71d482ab5c66c0d97429a6))
* enhance doctor diagnostics and tooling setup ([e852a97](https://github.com/royisme/BobaMixer/commit/e852a976b35e030391ecc9a9131719d5a4946ecd))
* implement comprehensive pricing.Load system with multi-source support ([fda8939](https://github.com/royisme/BobaMixer/commit/fda8939083c62ea10b80af567a9934bbd62e87a7))
* implement core execution features (batch 2) ([31d47fe](https://github.com/royisme/BobaMixer/commit/31d47fed6605a5cd52b95ca2e4a9f9c85b2b7018))
* implement DSL conditions, exploration mode, and feature flags ([2297a21](https://github.com/royisme/BobaMixer/commit/2297a2175d948e27ecee82dc0ad5d0f5cdc87936))
* implement httpx adapter and align stats APIs (M2-M3) ([85bfb4d](https://github.com/royisme/BobaMixer/commit/85bfb4dd8c7808f66404ca64254aeb29ed9b5752))
* implement structured logging and connect TUI dashboard ([5715034](https://github.com/royisme/BobaMixer/commit/5715034c8792521da69ed515590ed4f6f3dc4c85)), closes [#P0-4](https://github.com/royisme/BobaMixer/issues/P0-4) [#P2-5](https://github.com/royisme/BobaMixer/issues/P2-5) [#P1-2](https://github.com/royisme/BobaMixer/issues/P1-2) [#P3-3](https://github.com/royisme/BobaMixer/issues/P3-3)
* implement TDD spec core modules (M1-M2 partial) ([b71fea3](https://github.com/royisme/BobaMixer/commit/b71fea3dcf714c54b8277aa0919ccff4a93a641c))
* prioritize remote pricing before cache ([404ee5e](https://github.com/royisme/BobaMixer/commit/404ee5e72381a55ba1e1540d2df4be563ebb4af7))



# [1.0.0](https://github.com/royisme/BobaMixer/compare/v1.0.8...v1.0.0) (2025-11-15)


### Bug Fixes

* correct parseRows implementation in suggestions store ([d895fd1](https://github.com/royisme/BobaMixer/commit/d895fd1aeeac90a410271f70f894a4b01e4df1c5))
* correct version progression in database migration ([e61a4e2](https://github.com/royisme/BobaMixer/commit/e61a4e2ddea784c46c6b53ba373acd6e9e8ba494))
* improve routing DSL and suggestion parsing ([3d43c3f](https://github.com/royisme/BobaMixer/commit/3d43c3fe1d167f9cbc52fa0ec820eb234278c744))
* prevent zero-priced models and dead links ([c898096](https://github.com/royisme/BobaMixer/commit/c898096a64a5b343fdfd6cbef0118b2885016934))
* resolve all linter warnings for CI compliance ([db18f82](https://github.com/royisme/BobaMixer/commit/db18f82dd645ebcb243311e8251275804ae06059))
* resolve golint and typecheck errors ([479576f](https://github.com/royisme/BobaMixer/commit/479576f392ea83689431de5d2717de2138710216))
* resolve routing ctx_chars matching and linter issues ([aa06eb2](https://github.com/royisme/BobaMixer/commit/aa06eb24059f2f10715ccd672ff4f3103d3e9453))
* respect pricing.yaml sources configuration in Load() ([c1e26a2](https://github.com/royisme/BobaMixer/commit/c1e26a2743b93e25daa64c3c529497a76f15018b))


### Features

* add boba doctor --pricing command and fix linter issues ([4be12c9](https://github.com/royisme/BobaMixer/commit/4be12c9b44014132c9c86aa104f678f788f646c7))
* add foundational logging, secrets, and db helpers ([5ac9286](https://github.com/royisme/BobaMixer/commit/5ac92861d46e2764c2488b0f676f9e0786cb9478))
* add HTTP retry and enhanced diagnostics (batch 3) ([b0691a9](https://github.com/royisme/BobaMixer/commit/b0691a9eda3d0fe5fb56c6679b7fd5a0ef870027))
* align routing package with TDD spec (M3 partial) ([cc2ec37](https://github.com/royisme/BobaMixer/commit/cc2ec37261d8b8a49d71d482ab5c66c0d97429a6))
* enhance doctor diagnostics and tooling setup ([e852a97](https://github.com/royisme/BobaMixer/commit/e852a976b35e030391ecc9a9131719d5a4946ecd))
* implement comprehensive pricing.Load system with multi-source support ([fda8939](https://github.com/royisme/BobaMixer/commit/fda8939083c62ea10b80af567a9934bbd62e87a7))
* implement core execution features (batch 2) ([31d47fe](https://github.com/royisme/BobaMixer/commit/31d47fed6605a5cd52b95ca2e4a9f9c85b2b7018))
* implement DSL conditions, exploration mode, and feature flags ([2297a21](https://github.com/royisme/BobaMixer/commit/2297a2175d948e27ecee82dc0ad5d0f5cdc87936))
* implement httpx adapter and align stats APIs (M2-M3) ([85bfb4d](https://github.com/royisme/BobaMixer/commit/85bfb4dd8c7808f66404ca64254aeb29ed9b5752))
* implement structured logging and connect TUI dashboard ([5715034](https://github.com/royisme/BobaMixer/commit/5715034c8792521da69ed515590ed4f6f3dc4c85)), closes [#P0-4](https://github.com/royisme/BobaMixer/issues/P0-4) [#P2-5](https://github.com/royisme/BobaMixer/issues/P2-5) [#P1-2](https://github.com/royisme/BobaMixer/issues/P1-2) [#P3-3](https://github.com/royisme/BobaMixer/issues/P3-3)
* implement TDD spec core modules (M1-M2 partial) ([b71fea3](https://github.com/royisme/BobaMixer/commit/b71fea3dcf714c54b8277aa0919ccff4a93a641c))
* prioritize remote pricing before cache ([404ee5e](https://github.com/royisme/BobaMixer/commit/404ee5e72381a55ba1e1540d2df4be563ebb4af7))



# [1.0.0](https://github.com/royisme/BobaMixer/compare/v1.0.8...v1.0.0) (2025-11-14)


### Bug Fixes

* correct parseRows implementation in suggestions store ([d895fd1](https://github.com/royisme/BobaMixer/commit/d895fd1aeeac90a410271f70f894a4b01e4df1c5))
* correct version progression in database migration ([e61a4e2](https://github.com/royisme/BobaMixer/commit/e61a4e2ddea784c46c6b53ba373acd6e9e8ba494))
* improve routing DSL and suggestion parsing ([3d43c3f](https://github.com/royisme/BobaMixer/commit/3d43c3fe1d167f9cbc52fa0ec820eb234278c744))
* resolve all linter warnings for CI compliance ([db18f82](https://github.com/royisme/BobaMixer/commit/db18f82dd645ebcb243311e8251275804ae06059))
* resolve golint and typecheck errors ([479576f](https://github.com/royisme/BobaMixer/commit/479576f392ea83689431de5d2717de2138710216))
* resolve routing ctx_chars matching and linter issues ([aa06eb2](https://github.com/royisme/BobaMixer/commit/aa06eb24059f2f10715ccd672ff4f3103d3e9453))


### Features

* add foundational logging, secrets, and db helpers ([5ac9286](https://github.com/royisme/BobaMixer/commit/5ac92861d46e2764c2488b0f676f9e0786cb9478))
* add HTTP retry and enhanced diagnostics (batch 3) ([b0691a9](https://github.com/royisme/BobaMixer/commit/b0691a9eda3d0fe5fb56c6679b7fd5a0ef870027))
* align routing package with TDD spec (M3 partial) ([cc2ec37](https://github.com/royisme/BobaMixer/commit/cc2ec37261d8b8a49d71d482ab5c66c0d97429a6))
* enhance doctor diagnostics and tooling setup ([e852a97](https://github.com/royisme/BobaMixer/commit/e852a976b35e030391ecc9a9131719d5a4946ecd))
* implement core execution features (batch 2) ([31d47fe](https://github.com/royisme/BobaMixer/commit/31d47fed6605a5cd52b95ca2e4a9f9c85b2b7018))
* implement DSL conditions, exploration mode, and feature flags ([2297a21](https://github.com/royisme/BobaMixer/commit/2297a2175d948e27ecee82dc0ad5d0f5cdc87936))
* implement httpx adapter and align stats APIs (M2-M3) ([85bfb4d](https://github.com/royisme/BobaMixer/commit/85bfb4dd8c7808f66404ca64254aeb29ed9b5752))
* implement structured logging and connect TUI dashboard ([5715034](https://github.com/royisme/BobaMixer/commit/5715034c8792521da69ed515590ed4f6f3dc4c85)), closes [#P0-4](https://github.com/royisme/BobaMixer/issues/P0-4) [#P2-5](https://github.com/royisme/BobaMixer/issues/P2-5) [#P1-2](https://github.com/royisme/BobaMixer/issues/P1-2) [#P3-3](https://github.com/royisme/BobaMixer/issues/P3-3)
* implement TDD spec core modules (M1-M2 partial) ([b71fea3](https://github.com/royisme/BobaMixer/commit/b71fea3dcf714c54b8277aa0919ccff4a93a641c))
* prioritize remote pricing before cache ([404ee5e](https://github.com/royisme/BobaMixer/commit/404ee5e72381a55ba1e1540d2df4be563ebb4af7))



# [1.0.0](https://github.com/royisme/BobaMixer/compare/v1.0.8...v1.0.0) (2025-11-14)


### Bug Fixes

* correct parseRows implementation in suggestions store ([d895fd1](https://github.com/royisme/BobaMixer/commit/d895fd1aeeac90a410271f70f894a4b01e4df1c5))
* correct version progression in database migration ([e61a4e2](https://github.com/royisme/BobaMixer/commit/e61a4e2ddea784c46c6b53ba373acd6e9e8ba494))
* improve routing DSL and suggestion parsing ([3d43c3f](https://github.com/royisme/BobaMixer/commit/3d43c3fe1d167f9cbc52fa0ec820eb234278c744))
* resolve all linter warnings for CI compliance ([db18f82](https://github.com/royisme/BobaMixer/commit/db18f82dd645ebcb243311e8251275804ae06059))
* resolve golint and typecheck errors ([479576f](https://github.com/royisme/BobaMixer/commit/479576f392ea83689431de5d2717de2138710216))
* resolve routing ctx_chars matching and linter issues ([aa06eb2](https://github.com/royisme/BobaMixer/commit/aa06eb24059f2f10715ccd672ff4f3103d3e9453))


### Features

* add foundational logging, secrets, and db helpers ([5ac9286](https://github.com/royisme/BobaMixer/commit/5ac92861d46e2764c2488b0f676f9e0786cb9478))
* add HTTP retry and enhanced diagnostics (batch 3) ([b0691a9](https://github.com/royisme/BobaMixer/commit/b0691a9eda3d0fe5fb56c6679b7fd5a0ef870027))
* align routing package with TDD spec (M3 partial) ([cc2ec37](https://github.com/royisme/BobaMixer/commit/cc2ec37261d8b8a49d71d482ab5c66c0d97429a6))
* enhance doctor diagnostics and tooling setup ([e852a97](https://github.com/royisme/BobaMixer/commit/e852a976b35e030391ecc9a9131719d5a4946ecd))
* implement core execution features (batch 2) ([31d47fe](https://github.com/royisme/BobaMixer/commit/31d47fed6605a5cd52b95ca2e4a9f9c85b2b7018))
* implement DSL conditions, exploration mode, and feature flags ([2297a21](https://github.com/royisme/BobaMixer/commit/2297a2175d948e27ecee82dc0ad5d0f5cdc87936))
* implement httpx adapter and align stats APIs (M2-M3) ([85bfb4d](https://github.com/royisme/BobaMixer/commit/85bfb4dd8c7808f66404ca64254aeb29ed9b5752))
* implement structured logging and connect TUI dashboard ([5715034](https://github.com/royisme/BobaMixer/commit/5715034c8792521da69ed515590ed4f6f3dc4c85)), closes [#P0-4](https://github.com/royisme/BobaMixer/issues/P0-4) [#P2-5](https://github.com/royisme/BobaMixer/issues/P2-5) [#P1-2](https://github.com/royisme/BobaMixer/issues/P1-2) [#P3-3](https://github.com/royisme/BobaMixer/issues/P3-3)
* implement TDD spec core modules (M1-M2 partial) ([b71fea3](https://github.com/royisme/BobaMixer/commit/b71fea3dcf714c54b8277aa0919ccff4a93a641c))
* prioritize remote pricing before cache ([404ee5e](https://github.com/royisme/BobaMixer/commit/404ee5e72381a55ba1e1540d2df4be563ebb4af7))



# [1.0.0](https://github.com/royisme/BobaMixer/compare/v1.0.8...v1.0.0) (2025-11-14)


### Bug Fixes

* correct parseRows implementation in suggestions store ([d895fd1](https://github.com/royisme/BobaMixer/commit/d895fd1aeeac90a410271f70f894a4b01e4df1c5))
* correct version progression in database migration ([e61a4e2](https://github.com/royisme/BobaMixer/commit/e61a4e2ddea784c46c6b53ba373acd6e9e8ba494))
* improve routing DSL and suggestion parsing ([3d43c3f](https://github.com/royisme/BobaMixer/commit/3d43c3fe1d167f9cbc52fa0ec820eb234278c744))
* resolve all linter warnings for CI compliance ([db18f82](https://github.com/royisme/BobaMixer/commit/db18f82dd645ebcb243311e8251275804ae06059))
* resolve golint and typecheck errors ([479576f](https://github.com/royisme/BobaMixer/commit/479576f392ea83689431de5d2717de2138710216))
* resolve routing ctx_chars matching and linter issues ([aa06eb2](https://github.com/royisme/BobaMixer/commit/aa06eb24059f2f10715ccd672ff4f3103d3e9453))


### Features

* add foundational logging, secrets, and db helpers ([5ac9286](https://github.com/royisme/BobaMixer/commit/5ac92861d46e2764c2488b0f676f9e0786cb9478))
* add HTTP retry and enhanced diagnostics (batch 3) ([b0691a9](https://github.com/royisme/BobaMixer/commit/b0691a9eda3d0fe5fb56c6679b7fd5a0ef870027))
* align routing package with TDD spec (M3 partial) ([cc2ec37](https://github.com/royisme/BobaMixer/commit/cc2ec37261d8b8a49d71d482ab5c66c0d97429a6))
* enhance doctor diagnostics and tooling setup ([e852a97](https://github.com/royisme/BobaMixer/commit/e852a976b35e030391ecc9a9131719d5a4946ecd))
* implement core execution features (batch 2) ([31d47fe](https://github.com/royisme/BobaMixer/commit/31d47fed6605a5cd52b95ca2e4a9f9c85b2b7018))
* implement DSL conditions, exploration mode, and feature flags ([2297a21](https://github.com/royisme/BobaMixer/commit/2297a2175d948e27ecee82dc0ad5d0f5cdc87936))
* implement httpx adapter and align stats APIs (M2-M3) ([85bfb4d](https://github.com/royisme/BobaMixer/commit/85bfb4dd8c7808f66404ca64254aeb29ed9b5752))
* implement structured logging and connect TUI dashboard ([5715034](https://github.com/royisme/BobaMixer/commit/5715034c8792521da69ed515590ed4f6f3dc4c85)), closes [#P0-4](https://github.com/royisme/BobaMixer/issues/P0-4) [#P2-5](https://github.com/royisme/BobaMixer/issues/P2-5) [#P1-2](https://github.com/royisme/BobaMixer/issues/P1-2) [#P3-3](https://github.com/royisme/BobaMixer/issues/P3-3)
* implement TDD spec core modules (M1-M2 partial) ([b71fea3](https://github.com/royisme/BobaMixer/commit/b71fea3dcf714c54b8277aa0919ccff4a93a641c))
* prioritize remote pricing before cache ([404ee5e](https://github.com/royisme/BobaMixer/commit/404ee5e72381a55ba1e1540d2df4be563ebb4af7))



# [1.0.0](https://github.com/royisme/BobaMixer/compare/v1.0.8...v1.0.0) (2025-11-14)


### Bug Fixes

* correct parseRows implementation in suggestions store ([d895fd1](https://github.com/royisme/BobaMixer/commit/d895fd1aeeac90a410271f70f894a4b01e4df1c5))
* correct version progression in database migration ([e61a4e2](https://github.com/royisme/BobaMixer/commit/e61a4e2ddea784c46c6b53ba373acd6e9e8ba494))
* improve routing DSL and suggestion parsing ([3d43c3f](https://github.com/royisme/BobaMixer/commit/3d43c3fe1d167f9cbc52fa0ec820eb234278c744))
* resolve all linter warnings for CI compliance ([db18f82](https://github.com/royisme/BobaMixer/commit/db18f82dd645ebcb243311e8251275804ae06059))
* resolve golint and typecheck errors ([479576f](https://github.com/royisme/BobaMixer/commit/479576f392ea83689431de5d2717de2138710216))
* resolve routing ctx_chars matching and linter issues ([aa06eb2](https://github.com/royisme/BobaMixer/commit/aa06eb24059f2f10715ccd672ff4f3103d3e9453))


### Features

* add foundational logging, secrets, and db helpers ([5ac9286](https://github.com/royisme/BobaMixer/commit/5ac92861d46e2764c2488b0f676f9e0786cb9478))
* add HTTP retry and enhanced diagnostics (batch 3) ([b0691a9](https://github.com/royisme/BobaMixer/commit/b0691a9eda3d0fe5fb56c6679b7fd5a0ef870027))
* align routing package with TDD spec (M3 partial) ([cc2ec37](https://github.com/royisme/BobaMixer/commit/cc2ec37261d8b8a49d71d482ab5c66c0d97429a6))
* enhance doctor diagnostics and tooling setup ([e852a97](https://github.com/royisme/BobaMixer/commit/e852a976b35e030391ecc9a9131719d5a4946ecd))
* implement core execution features (batch 2) ([31d47fe](https://github.com/royisme/BobaMixer/commit/31d47fed6605a5cd52b95ca2e4a9f9c85b2b7018))
* implement DSL conditions, exploration mode, and feature flags ([2297a21](https://github.com/royisme/BobaMixer/commit/2297a2175d948e27ecee82dc0ad5d0f5cdc87936))
* implement httpx adapter and align stats APIs (M2-M3) ([85bfb4d](https://github.com/royisme/BobaMixer/commit/85bfb4dd8c7808f66404ca64254aeb29ed9b5752))
* implement structured logging and connect TUI dashboard ([5715034](https://github.com/royisme/BobaMixer/commit/5715034c8792521da69ed515590ed4f6f3dc4c85)), closes [#P0-4](https://github.com/royisme/BobaMixer/issues/P0-4) [#P2-5](https://github.com/royisme/BobaMixer/issues/P2-5) [#P1-2](https://github.com/royisme/BobaMixer/issues/P1-2) [#P3-3](https://github.com/royisme/BobaMixer/issues/P3-3)
* implement TDD spec core modules (M1-M2 partial) ([b71fea3](https://github.com/royisme/BobaMixer/commit/b71fea3dcf714c54b8277aa0919ccff4a93a641c))
* prioritize remote pricing before cache ([404ee5e](https://github.com/royisme/BobaMixer/commit/404ee5e72381a55ba1e1540d2df4be563ebb4af7))



# [1.0.0](https://github.com/royisme/BobaMixer/compare/v1.0.8...v1.0.0) (2025-11-14)


### Bug Fixes

* correct parseRows implementation in suggestions store ([d895fd1](https://github.com/royisme/BobaMixer/commit/d895fd1aeeac90a410271f70f894a4b01e4df1c5))
* correct version progression in database migration ([e61a4e2](https://github.com/royisme/BobaMixer/commit/e61a4e2ddea784c46c6b53ba373acd6e9e8ba494))
* improve routing DSL and suggestion parsing ([3d43c3f](https://github.com/royisme/BobaMixer/commit/3d43c3fe1d167f9cbc52fa0ec820eb234278c744))
* resolve golint and typecheck errors ([479576f](https://github.com/royisme/BobaMixer/commit/479576f392ea83689431de5d2717de2138710216))


### Features

* add foundational logging, secrets, and db helpers ([5ac9286](https://github.com/royisme/BobaMixer/commit/5ac92861d46e2764c2488b0f676f9e0786cb9478))
* add HTTP retry and enhanced diagnostics (batch 3) ([b0691a9](https://github.com/royisme/BobaMixer/commit/b0691a9eda3d0fe5fb56c6679b7fd5a0ef870027))
* enhance doctor diagnostics and tooling setup ([e852a97](https://github.com/royisme/BobaMixer/commit/e852a976b35e030391ecc9a9131719d5a4946ecd))
* implement core execution features (batch 2) ([31d47fe](https://github.com/royisme/BobaMixer/commit/31d47fed6605a5cd52b95ca2e4a9f9c85b2b7018))
* implement DSL conditions, exploration mode, and feature flags ([2297a21](https://github.com/royisme/BobaMixer/commit/2297a2175d948e27ecee82dc0ad5d0f5cdc87936))
* implement structured logging and connect TUI dashboard ([5715034](https://github.com/royisme/BobaMixer/commit/5715034c8792521da69ed515590ed4f6f3dc4c85)), closes [#P0-4](https://github.com/royisme/BobaMixer/issues/P0-4) [#P2-5](https://github.com/royisme/BobaMixer/issues/P2-5) [#P1-2](https://github.com/royisme/BobaMixer/issues/P1-2) [#P3-3](https://github.com/royisme/BobaMixer/issues/P3-3)
* prioritize remote pricing before cache ([404ee5e](https://github.com/royisme/BobaMixer/commit/404ee5e72381a55ba1e1540d2df4be563ebb4af7))



# [1.0.0](https://github.com/royisme/BobaMixer/compare/v1.0.8...v1.0.0) (2025-11-14)


### Bug Fixes

* correct parseRows implementation in suggestions store ([d895fd1](https://github.com/royisme/BobaMixer/commit/d895fd1aeeac90a410271f70f894a4b01e4df1c5))
* correct version progression in database migration ([e61a4e2](https://github.com/royisme/BobaMixer/commit/e61a4e2ddea784c46c6b53ba373acd6e9e8ba494))
* improve routing DSL and suggestion parsing ([3d43c3f](https://github.com/royisme/BobaMixer/commit/3d43c3fe1d167f9cbc52fa0ec820eb234278c744))
* resolve golint and typecheck errors ([479576f](https://github.com/royisme/BobaMixer/commit/479576f392ea83689431de5d2717de2138710216))


### Features

* add HTTP retry and enhanced diagnostics (batch 3) ([b0691a9](https://github.com/royisme/BobaMixer/commit/b0691a9eda3d0fe5fb56c6679b7fd5a0ef870027))
* enhance doctor diagnostics and tooling setup ([e852a97](https://github.com/royisme/BobaMixer/commit/e852a976b35e030391ecc9a9131719d5a4946ecd))
* implement core execution features (batch 2) ([31d47fe](https://github.com/royisme/BobaMixer/commit/31d47fed6605a5cd52b95ca2e4a9f9c85b2b7018))
* implement DSL conditions, exploration mode, and feature flags ([2297a21](https://github.com/royisme/BobaMixer/commit/2297a2175d948e27ecee82dc0ad5d0f5cdc87936))
* implement structured logging and connect TUI dashboard ([5715034](https://github.com/royisme/BobaMixer/commit/5715034c8792521da69ed515590ed4f6f3dc4c85)), closes [#P0-4](https://github.com/royisme/BobaMixer/issues/P0-4) [#P2-5](https://github.com/royisme/BobaMixer/issues/P2-5) [#P1-2](https://github.com/royisme/BobaMixer/issues/P1-2) [#P3-3](https://github.com/royisme/BobaMixer/issues/P3-3)
* prioritize remote pricing before cache ([404ee5e](https://github.com/royisme/BobaMixer/commit/404ee5e72381a55ba1e1540d2df4be563ebb4af7))



# [1.0.0](https://github.com/royisme/BobaMixer/compare/v1.0.8...v1.0.0) (2025-11-14)


### Bug Fixes

* correct parseRows implementation in suggestions store ([d895fd1](https://github.com/royisme/BobaMixer/commit/d895fd1aeeac90a410271f70f894a4b01e4df1c5))
* correct version progression in database migration ([e61a4e2](https://github.com/royisme/BobaMixer/commit/e61a4e2ddea784c46c6b53ba373acd6e9e8ba494))
* improve routing DSL and suggestion parsing ([3d43c3f](https://github.com/royisme/BobaMixer/commit/3d43c3fe1d167f9cbc52fa0ec820eb234278c744))
* resolve golint and typecheck errors ([479576f](https://github.com/royisme/BobaMixer/commit/479576f392ea83689431de5d2717de2138710216))


### Features

* add HTTP retry and enhanced diagnostics (batch 3) ([b0691a9](https://github.com/royisme/BobaMixer/commit/b0691a9eda3d0fe5fb56c6679b7fd5a0ef870027))
* enhance doctor diagnostics and tooling setup ([e852a97](https://github.com/royisme/BobaMixer/commit/e852a976b35e030391ecc9a9131719d5a4946ecd))
* implement core execution features (batch 2) ([31d47fe](https://github.com/royisme/BobaMixer/commit/31d47fed6605a5cd52b95ca2e4a9f9c85b2b7018))
* implement DSL conditions, exploration mode, and feature flags ([2297a21](https://github.com/royisme/BobaMixer/commit/2297a2175d948e27ecee82dc0ad5d0f5cdc87936))
* implement structured logging and connect TUI dashboard ([5715034](https://github.com/royisme/BobaMixer/commit/5715034c8792521da69ed515590ed4f6f3dc4c85)), closes [#P0-4](https://github.com/royisme/BobaMixer/issues/P0-4) [#P2-5](https://github.com/royisme/BobaMixer/issues/P2-5) [#P1-2](https://github.com/royisme/BobaMixer/issues/P1-2) [#P3-3](https://github.com/royisme/BobaMixer/issues/P3-3)



# [1.0.0](https://github.com/royisme/BobaMixer/compare/v1.0.8...v1.0.0) (2025-11-14)


### Bug Fixes

* correct parseRows implementation in suggestions store ([d895fd1](https://github.com/royisme/BobaMixer/commit/d895fd1aeeac90a410271f70f894a4b01e4df1c5))
* correct version progression in database migration ([e61a4e2](https://github.com/royisme/BobaMixer/commit/e61a4e2ddea784c46c6b53ba373acd6e9e8ba494))
* improve routing DSL and suggestion parsing ([3d43c3f](https://github.com/royisme/BobaMixer/commit/3d43c3fe1d167f9cbc52fa0ec820eb234278c744))
* resolve golint and typecheck errors ([479576f](https://github.com/royisme/BobaMixer/commit/479576f392ea83689431de5d2717de2138710216))


### Features

* add HTTP retry and enhanced diagnostics (batch 3) ([b0691a9](https://github.com/royisme/BobaMixer/commit/b0691a9eda3d0fe5fb56c6679b7fd5a0ef870027))
* enhance doctor diagnostics and tooling setup ([e852a97](https://github.com/royisme/BobaMixer/commit/e852a976b35e030391ecc9a9131719d5a4946ecd))
* implement core execution features (batch 2) ([31d47fe](https://github.com/royisme/BobaMixer/commit/31d47fed6605a5cd52b95ca2e4a9f9c85b2b7018))
* implement DSL conditions, exploration mode, and feature flags ([2297a21](https://github.com/royisme/BobaMixer/commit/2297a2175d948e27ecee82dc0ad5d0f5cdc87936))
* implement structured logging and connect TUI dashboard ([5715034](https://github.com/royisme/BobaMixer/commit/5715034c8792521da69ed515590ed4f6f3dc4c85)), closes [#P0-4](https://github.com/royisme/BobaMixer/issues/P0-4) [#P2-5](https://github.com/royisme/BobaMixer/issues/P2-5) [#P1-2](https://github.com/royisme/BobaMixer/issues/P1-2) [#P3-3](https://github.com/royisme/BobaMixer/issues/P3-3)



# [1.0.0](https://github.com/royisme/BobaMixer/compare/v1.0.8...v1.0.0) (2025-11-14)


### Bug Fixes

* correct parseRows implementation in suggestions store ([d895fd1](https://github.com/royisme/BobaMixer/commit/d895fd1aeeac90a410271f70f894a4b01e4df1c5))
* correct version progression in database migration ([e61a4e2](https://github.com/royisme/BobaMixer/commit/e61a4e2ddea784c46c6b53ba373acd6e9e8ba494))
* improve routing DSL and suggestion parsing ([3d43c3f](https://github.com/royisme/BobaMixer/commit/3d43c3fe1d167f9cbc52fa0ec820eb234278c744))
* resolve golint and typecheck errors ([479576f](https://github.com/royisme/BobaMixer/commit/479576f392ea83689431de5d2717de2138710216))


### Features

* add HTTP retry and enhanced diagnostics (batch 3) ([b0691a9](https://github.com/royisme/BobaMixer/commit/b0691a9eda3d0fe5fb56c6679b7fd5a0ef870027))
* enhance doctor diagnostics and tooling setup ([e852a97](https://github.com/royisme/BobaMixer/commit/e852a976b35e030391ecc9a9131719d5a4946ecd))
* implement core execution features (batch 2) ([31d47fe](https://github.com/royisme/BobaMixer/commit/31d47fed6605a5cd52b95ca2e4a9f9c85b2b7018))
* implement DSL conditions, exploration mode, and feature flags ([2297a21](https://github.com/royisme/BobaMixer/commit/2297a2175d948e27ecee82dc0ad5d0f5cdc87936))
* implement structured logging and connect TUI dashboard ([5715034](https://github.com/royisme/BobaMixer/commit/5715034c8792521da69ed515590ed4f6f3dc4c85)), closes [#P0-4](https://github.com/royisme/BobaMixer/issues/P0-4) [#P2-5](https://github.com/royisme/BobaMixer/issues/P2-5) [#P1-2](https://github.com/royisme/BobaMixer/issues/P1-2) [#P3-3](https://github.com/royisme/BobaMixer/issues/P3-3)



# [1.0.0](https://github.com/royisme/BobaMixer/compare/v1.0.8...v1.0.0) (2025-11-14)


### Bug Fixes

* correct parseRows implementation in suggestions store ([d895fd1](https://github.com/royisme/BobaMixer/commit/d895fd1aeeac90a410271f70f894a4b01e4df1c5))
* correct version progression in database migration ([e61a4e2](https://github.com/royisme/BobaMixer/commit/e61a4e2ddea784c46c6b53ba373acd6e9e8ba494))
* improve routing DSL and suggestion parsing ([3d43c3f](https://github.com/royisme/BobaMixer/commit/3d43c3fe1d167f9cbc52fa0ec820eb234278c744))
* resolve golint and typecheck errors ([479576f](https://github.com/royisme/BobaMixer/commit/479576f392ea83689431de5d2717de2138710216))


### Features

* add HTTP retry and enhanced diagnostics (batch 3) ([b0691a9](https://github.com/royisme/BobaMixer/commit/b0691a9eda3d0fe5fb56c6679b7fd5a0ef870027))
* enhance doctor diagnostics and tooling setup ([e852a97](https://github.com/royisme/BobaMixer/commit/e852a976b35e030391ecc9a9131719d5a4946ecd))
* implement core execution features (batch 2) ([31d47fe](https://github.com/royisme/BobaMixer/commit/31d47fed6605a5cd52b95ca2e4a9f9c85b2b7018))
* implement DSL conditions, exploration mode, and feature flags ([2297a21](https://github.com/royisme/BobaMixer/commit/2297a2175d948e27ecee82dc0ad5d0f5cdc87936))
* implement structured logging and connect TUI dashboard ([5715034](https://github.com/royisme/BobaMixer/commit/5715034c8792521da69ed515590ed4f6f3dc4c85)), closes [#P0-4](https://github.com/royisme/BobaMixer/issues/P0-4) [#P2-5](https://github.com/royisme/BobaMixer/issues/P2-5) [#P1-2](https://github.com/royisme/BobaMixer/issues/P1-2) [#P3-3](https://github.com/royisme/BobaMixer/issues/P3-3)



# [1.0.0](https://github.com/royisme/BobaMixer/compare/v1.0.8...v1.0.0) (2025-11-14)


### Features

* add HTTP retry and enhanced diagnostics (batch 3) ([b0691a9](https://github.com/royisme/BobaMixer/commit/b0691a9eda3d0fe5fb56c6679b7fd5a0ef870027))
* enhance doctor diagnostics and tooling setup ([e852a97](https://github.com/royisme/BobaMixer/commit/e852a976b35e030391ecc9a9131719d5a4946ecd))
* implement core execution features (batch 2) ([31d47fe](https://github.com/royisme/BobaMixer/commit/31d47fed6605a5cd52b95ca2e4a9f9c85b2b7018))
* implement structured logging and connect TUI dashboard ([5715034](https://github.com/royisme/BobaMixer/commit/5715034c8792521da69ed515590ed4f6f3dc4c85)), closes [#P0-4](https://github.com/royisme/BobaMixer/issues/P0-4) [#P2-5](https://github.com/royisme/BobaMixer/issues/P2-5) [#P1-2](https://github.com/royisme/BobaMixer/issues/P1-2) [#P3-3](https://github.com/royisme/BobaMixer/issues/P3-3)



# [1.0.0](https://github.com/royisme/BobaMixer/compare/v1.0.8...v1.0.0) (2025-11-14)


### Features

* add HTTP retry and enhanced diagnostics (batch 3) ([b0691a9](https://github.com/royisme/BobaMixer/commit/b0691a9eda3d0fe5fb56c6679b7fd5a0ef870027))
* implement core execution features (batch 2) ([31d47fe](https://github.com/royisme/BobaMixer/commit/31d47fed6605a5cd52b95ca2e4a9f9c85b2b7018))
* implement structured logging and connect TUI dashboard ([5715034](https://github.com/royisme/BobaMixer/commit/5715034c8792521da69ed515590ed4f6f3dc4c85)), closes [#P0-4](https://github.com/royisme/BobaMixer/issues/P0-4) [#P2-5](https://github.com/royisme/BobaMixer/issues/P2-5) [#P1-2](https://github.com/royisme/BobaMixer/issues/P1-2) [#P3-3](https://github.com/royisme/BobaMixer/issues/P3-3)



# [1.0.0](https://github.com/royisme/BobaMixer/compare/v1.0.7...v1.0.0) (2025-11-14)



# [1.0.0](https://github.com/royisme/BobaMixer/compare/v1.0.6...v1.0.0) (2025-11-14)



# [1.0.0](https://github.com/royisme/BobaMixer/compare/v1.0.6...v1.0.0) (2025-11-14)



# [1.0.0](https://github.com/royisme/BobaMixer/compare/v1.0.6...v1.0.0) (2025-11-14)



# [1.0.0](https://github.com/royisme/BobaMixer/compare/v1.0.5...v1.0.0) (2025-11-14)



# [1.0.0](https://github.com/royisme/BobaMixer/compare/v0.1.0...v1.0.0) (2025-11-13)


### Features

* add complete release workflow with goreleaser and conventional commits ([3471f32](https://github.com/royisme/BobaMixer/commit/3471f32ba9ec6ca76357f84bd358f05e0935c0b4))



# Changelog

All notable changes to BobaMixer will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### 🚀 Features
- *Feature descriptions here*

### 🐛 Bug Fixes  
- *Bug fix descriptions here*

### 🔧 Improvements
- *Improvement descriptions here*

### 📚 Documentation
- *Documentation changes here*

### ⚠️ Breaking Changes
- *Breaking change descriptions here*

---

## [1.0.0] - 2024-11-13

### 🚀 Features
- **Intelligent AI Routing**: Automatic provider selection based on context, cost, and performance
- **Multi-Provider Support**: Unified interface for OpenAI, Anthropic Claude, and local models
- **Budget Management**: Multi-level budget tracking (global, project, profile) with real-time alerts
- **Usage Analytics**: Comprehensive token count, cost tracking, and performance metrics
- **Project Awareness**: Git integration for automatic project and branch context detection
- **TUI Dashboard**: Beautiful terminal UI for real-time monitoring and analytics
- **Cost Optimization**: Intelligent suggestions for reducing AI costs
- **MCP Support**: Model Context Protocol integration for advanced AI workflows

### 🛠️ Configuration
- YAML-based configuration system with profiles, routes, and secrets
- Dynamic routing rules with DSL-based matching
- Project-specific configuration overrides
- Secure API key management with file permissions

### 📚 Documentation
- Comprehensive user documentation with getting started guide
- Configuration reference with practical examples
- API documentation and command reference

### 🔧 Technical
- Clean architecture with domain-driven design
- SQLite database for usage tracking
- Comprehensive test coverage
- Multi-platform binary distribution

### 🎯 Use Cases
- **Developers**: Intelligent code analysis and optimization suggestions
- **Teams**: Budget management and usage analytics across organizations
- **Cost-Conscious Users**: Automatic routing to most cost-effective providers
- **Power Users**: Advanced configuration and customization options

---

## Version Guide

### Version Format
- **Major**: Breaking changes (X.0.0)
- **Minor**: New features (X.Y.0)  
- **Patch**: Bug fixes (X.Y.Z)

### Release Types
- **Stable**: Production-ready releases
- **Pre-release**: Alpha/beta for testing (e.g., 1.2.0-beta.1)

### Upgrade Guide
Major releases will include migration guides in the documentation.