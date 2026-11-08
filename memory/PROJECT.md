# Project memory

## Session 1 — 2026-11-09 — app built, instructions added

**Done:**

- Boki-WebApps is a Go GTK 4 app (gotk4, no libadwaita, no WebKit). It scans desktop files for browsers, writes `WebApp-*.desktop` launchers, and starts them with `gio launch`. Firefox can show a navbar. Chromium-family, including Brave Origin, uses `--app`.
- Theme CSS is only `~/.config/boki-webapps/theme.css` (panel background `#1d283a`, title bar `#6daea8`, title ink `#1d283a`, text `#c9e79a`). Icons, including custom pictures, live in `~/.local/share/boki-webapps/icons/`. The theme changer can restyle the CSS. It must not own the icons.
- List rows show the site icon, then the name. Remove and the window close are a text `×`. The search field uses the panel colors. Sizes: row icon 36px, label 18px, edit and remove 36px, add button 52px.
- A new app does not start from the browser icon. The URL lookup fills a site icon. Choose lists pictures already in the icons folder, then icons from the site.
- `WebApp-Youtube6921.desktop` launches with Brave Origin. Leave that file alone. `~/.local/bin/brave-basic` still points at a missing `/usr/bin/brave`. The scanner lists it. Do not hide it unless asked.
- `go test ./...` and `go build` succeed. The `free` cgo warnings are from gotk4's generated wrappers.
- `push.sh` and `set-git-cred.sh` point at `git@github.com:bojanstrkovski-21/boki-webapps.git`. `set-git-cred.sh` was not run. Repo is `main` at `850b2fc` (initial commit: code and both scripts), tracking `origin/main`.
- On "end session" / "end-session", write `memory/PROJECT.md` and sync `~/.agents/memory/` without asking, then report what was written. Same outcome as `~/.agents/commands/end-session.md` steps 4 and 6. dwm-quickshell was not edited.

**Open:**

- `CLAUDE.md` and `memory/` are uncommitted. Not pushed.
- No `.gitignore` yet. Add one only if asked.
