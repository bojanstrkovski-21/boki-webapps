# Boki-WebApps

A small desktop app that turns a website into a launcher. You give it a name and a URL. It writes a menu entry and opens the site in a browser you already have installed. It does not embed a browser.

Version 0.1.0. Application id `com.boki.WebApps`.

## What it does

- Finds installed browsers by reading their `.desktop` files. There is no hardcoded browser list.
- Saves each web app as `~/.local/share/applications/WebApp-<codename>.desktop`.
- Keeps a separate browser profile under `~/.local/share/ice/` when the web app is isolated.
- Picks a site icon from the page. You can replace it with a picture of your own.
- Firefox-family browsers can show a navigation bar. Chromium-family browsers, Brave Origin included, open with `--app` and have no back/forward bar in this app.
- Colors come from `~/.config/boki-webapps/theme.css`, so a theme changer can restyle the window. Icons stay in `~/.local/share/boki-webapps/icons/`.

## Requirements

- Linux with GTK 4, GLib, and `gio`
- Go 1.27 or newer
- A desktop browser (Firefox, Chromium, Brave Origin, Vivaldi, Epiphany, Falkon, and the same family)

## Build and run

```bash
cd ~/Projects/boki-webapps
go test ./...
go build -o /tmp/boki-webapps .
/tmp/boki-webapps
```

The app is not installed system-wide. The binary can stay in `/tmp` or anywhere you run it from. The launchers it writes are what show up in the menu.

## Where things live

| What | Where |
| --- | --- |
| Theme | `~/.config/boki-webapps/theme.css` |
| Icons | `~/.local/share/boki-webapps/icons/` |
| Launchers | `~/.local/share/applications/WebApp-*.desktop` |
| Firefox profiles | `~/.local/share/ice/firefox/<codename>/` |
| Chromium profiles | `~/.local/share/ice/profiles/<codename>/` |

`push.sh` and `set-git-cred.sh` are for you to run from your own terminal. `set-git-cred.sh` changes the global git identity.

## Credits

The launcher idea and the on-disk layout were studied from [webapp-manager](https://github.com/linuxmint/webapp-manager) by the Linux Mint team (GPL-3.0-or-later). That is the `WebApp-*.desktop` name, the profile directory `~/.local/share/ice/`, the Firefox navigation-bar switch, and Chromium’s `--app` window. This repository does not include their Python program or their CSS. The Go code here is separate.

When webapp-manager is installed, the first save of a Firefox web app copies `/usr/share/webapp-manager/firefox/profile` into that web app’s profile. The toolbar CSS in that profile is from [firefox-csshacks](https://github.com/MrOtherGuy/firefox-csshacks) by MrOtherGuy, Mozilla Public License 2.0. The copy keeps that notice. If that directory is not installed, this app writes its own short `user.js` and a small navbar rule instead.

The window is GTK 4 through [gotk4](https://github.com/diamondburned/gotk4) by diamondburned (Mozilla Public License 2.0). GTK itself is the GNOME toolkit.

`push.sh` and `set-git-cred.sh` are this repo’s copies of the scripts in [bole-labs-ai](https://github.com/bojanstrkovski-21/bole-labs-ai) (`~/.agents`), which took them from [session-defaults](https://github.com/bojanstrkovski-21/session-defaults).

## License

[MIT](LICENSE). Copyright (c) 2026 bojanstrkovski-21. Code and CSS from the projects above keep their own license and copyright notices. The Firefox profile copied at runtime is not covered by this MIT license.
