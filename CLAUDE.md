# Boki-WebApps

## Overview
Desktop app that turns a site into a menu launcher. The user names the app
and pastes a URL. The app writes a `WebApp-*.desktop` file and opens that
site in a browser that is already installed. It does not embed a browser.

Source is `/home/bojanstrko/Projects/boki-webapps`. A session may be rooted
at `~/Projects/yay`. Work for this app stays in this directory.

App id `com.boki.WebApps`. Version `0.1.0`.

## Stack
- **Language**: Go 1.27.3, module `boki-webapps`. `main.go` calls `ui.Run()`.
- **UI**: GTK 4 through `github.com/diamondburned/gotk4/pkg` (cgo, system GTK). Plain GTK. No libadwaita. No WebKit.
- **Packages**:
  - `internal/browser` finds browsers by reading `.desktop` files (user applications, Flatpak, `XDG_DATA_DIRS`, system). No hardcoded browser list. A file counts when `Categories` has `WebBrowser` or `MimeType` handles `http`/`https`, and the executable exists. Skip `Hidden`/`NoDisplay`, `WebApp-*`, `X-WebApp-URL`, Desktop Actions, and an `Exec` that contains `--app=` or `--application-mode`.
  - `internal/webapp` reads and writes `~/.local/share/applications/WebApp-<codename>.desktop` and profiles under `~/.local/share/ice/`. Codename is four digits. Launch is `gio launch`. Firefox-family copies a profile on first create and can show a navbar through `userChrome.css`. Chromium-family, Brave Origin included, uses `--app` and has no back/forward bar.
  - `internal/icon` fetches a site icon from the page and saves it under `~/.local/share/boki-webapps/icons/`. Choose lists pictures already in that folder first. On save, an absolute icon path is copied to `<codename><ext>` there.
  - `internal/theme` loads `~/.config/boki-webapps/theme.css` when it exists, and creates it from `DefaultCSS` only when it is missing. The user's theme changer owns the live file.
  - `internal/ui` is the window. New apps do not prefill a browser icon. A URL change fills the best site icon. The icon form field is `iconEntry` so it does not shadow package `icon`. GTK file filters use `AddMIMEType`.

## Paths
| What | Where |
| --- | --- |
| Source | `~/Projects/boki-webapps` |
| Theme CSS | `~/.config/boki-webapps/theme.css` only |
| Icons | `~/.local/share/boki-webapps/icons/` |
| Launchers | `~/.local/share/applications/WebApp-*.desktop` |
| Firefox profiles | `~/.local/share/ice/firefox/<codename>/` |
| Chromium profiles | `~/.local/share/ice/profiles/<codename>/` |
| Remote | `git@github.com:bojanstrkovski-21/boki-webapps.git` |

`~/.config/boki-webapps` is the theme engine's directory. Icons do not go there.

Colors that match the panel (archboki-dark): window background `#1d283a`, text `#c9e79a`, title bar `#6daea8`, title-bar ink `#1d283a`. The title bar is a GTK 4 header bar with native title buttons off. Close and row-remove are a text `×`, because the icon theme draws `window-close-symbolic` as an X inside a circle.

## Conventions
- `push.sh` and `set-git-cred.sh` are this repo's filled-in copies of the scripts in `~/.agents` (`project=boki-webapps`). They are for the user to run. See `~/.agents/README.md`.
- Check binaries go in `/tmp`, not in the tree.

## Constraints / Hard Rules
Global rules from `~/.agents` (confirm-first, no git unless asked, no force push) apply and are not
repeated here. Project rules:
- No libadwaita, no WebKit, no in-app back/forward bar, no system-wide install.
- This repository is MIT (`LICENSE`, copyright 2026 bojanstrkovski-21). Do not vendor webapp-manager's CSS or its GPL Python. The Firefox profile copied at runtime from an installed webapp-manager keeps that package's license and the firefox-csshacks MPL notice.
- Do not edit the theme changer (`~/.config/chadboki-qswm/quickshell-scripts/dwm-quickshell-theme`) or the theme engine. Edit `theme.css` only when the user asks for a color change, and keep `DefaultCSS` in `internal/theme/theme.go` in step with that file.
- Do not edit an existing `WebApp-*.desktop` unless the user asks for that file. `WebApp-Youtube6921.desktop` launches with Brave Origin. Leave it.
- `~/.local/bin/brave-basic` calls `/usr/bin/brave`, which is not installed. The installed browser is `brave-origin`. The scanner still lists `brave-basic` because the script file exists. Do not hide it unless asked. `gio launch` can return 0 after a wrapper has already failed.
- cgo warnings `conflicting types for built-in function 'free'` come from gotk4's generated wrappers. The build still succeeds. Do not patch the module cache.
- Never edit a file listed under **Frozen Files** below without explicit instruction naming that file.
- Never write credentials, tokens, or secrets into any file in this repo — not even a scratch/junk file.

## Frozen Files
- `~/.config/chadboki-qswm/quickshell-scripts/dwm-quickshell-theme` and the rest of the theme engine.
- `~/.local/share/applications/WebApp-Youtube6921.desktop` unless the user names that file.
- Anything under `~/Projects/yay` while working on this app.

## How to Run
```bash
cd ~/Projects/boki-webapps
go test ./...
go build -o /tmp/boki-webapps .
/tmp/boki-webapps
```

`BOKI_SMOKE=1` quits as soon as the window is up. `BOKI_LIVE_SCAN=1` and `BOKI_LIVE_ICON=1` enable the live browser and icon tests.

## Current state
The latest session is the top entry of `memory/PROJECT.md`; older sessions are in
`memory/archive/`.

## Session Workflow

Follows `~/.agents/commands/start-session.md`/`end-session.md`'s shape
(shared across Claude/Codex/Copilot per `~/.agents/README.md`), with this
project's own steps layered on top rather than replacing it. Same layering
as `~/Projects/dwm-quickshell/CLAUDE.md`. This project has no start-session
sync script and no `to-do-tasks.txt`.

- When the user says **"start session"** or **"start session - `<topic>`"**:
  1. Read `memory/PROJECT.md` (live sessions only; older ones are archived in `memory/archive/`).
  2. Show a compact summary: what the project is, current state, next steps.
  3. If a topic was given, confirm that's the focus; otherwise ask what the goal is for this session.
- When the user says **"end session"** or **"end-session"**:
  Follow `~/.agents/commands/end-session.md`. Its step 2 says "propose"; steps 4 and 6, and the Rules section, say the docs and the global memory are written without asking (2026-10-25). This project uses that: write, then report. Do not ask for permission to write memory.
  1. Summarize what was accomplished this session (3–5 bullets).
  2. Write a new `## Session N: <title>` entry at the top of `memory/PROJECT.md` (what was done, what is left). When the live file holds more than about 6 sessions, move the oldest into `memory/archive/`.
  3. If this session changed a rule in this file, edit it.
  4. **Check git state**: run `git status` (read-only) and compare against the remote if one's configured. If there are uncommitted or committed-but-unpushed changes, say so explicitly rather than staying silent — this is a reminder, not an action; never commit or push as part of ending a session.
  5. If a durable cross-project fact came up, write it to `~/.agents/memory/` and add or update its `MEMORY.md` index line. Check the index first. Update an existing fact. If nothing new qualifies, skip it. Do not ask either way. Detail that only this app needs stays in `memory/PROJECT.md`.
  6. Report what was written. Code and git still need an explicit ask.
  7. Do not run any git commands as part of ending a session.
