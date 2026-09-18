# Tasks: move-config-into-dir

## 1. Config path change

- [x] 1.1 Change `config.Path()` to return the XDG-aware `graft/config.toml` location and update the missing-config error message in `Load()` to identify the new path; verify unit tests cover the new path under both default and `XDG_CONFIG_HOME` scenarios and the missing-file error text

## 2. Test harness and docs

- [x] 2.1 Update `internal/e2e/harness.go` `WriteConfig` to write `config.toml` under the `graft/` subdirectory; verify all existing e2e tests still pass unchanged (`go test ./internal/e2e/...`)
- [x] 2.2 Update `README.md` to show the new config location, the `~/.config/graft/` directory layout with `hooks/` for collocated hook scripts, and the legacy-path migration note (`mv ~/.config/graft.toml ~/.config/graft/config.toml`); verify the documented commands work against the compiled binary

## 3. Verification

- [x] 3.1 Run `nix run .#check` and confirm it passes
