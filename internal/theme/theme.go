package theme

import (
	"os"
	"path/filepath"
)

const FileName = "theme.css"

const DefaultCSS = `/* Boki-WebApps theme.
   The theme changer can replace this file. The open window reloads it. */

window,
scrolledwindow,
viewport,
listbox {
  background-color: #1d283a;
  color: #c9e79a;
}

headerbar {
  background-color: #6daea8;
  color: #1d283a;
  min-height: 42px;
}

headerbar label,
headerbar button {
  color: #1d283a;
}

headerbar .app-title {
  font-weight: 600;
}

entry.search {
  background-color: #243044;
  color: #c9e79a;
  caret-color: #c9e79a;
  border: 1px solid #3d5c78;
  border-radius: 8px;
  min-height: 36px;
}

entry.search:focus {
  border-color: #6daea8;
}

entry.search image,
entry.search placeholder {
  color: #9eb8a4;
}

headerbar button.window-x {
  min-width: 28px;
  min-height: 28px;
  padding: 0 6px;
  margin-right: 6px;
  font-size: 20px;
  font-weight: 700;
  color: #1d283a;
  background: none;
  border: none;
  box-shadow: none;
}

listbox, listbox row {
  background-color: transparent;
}

listbox row {
  border-bottom: 2px solid #3d7ea6;
  padding: 6px 4px;
}

popover contents modelbutton,
popover label {
  color: #c9e79a;
}

button.flat {
  background: none;
  box-shadow: none;
  border: none;
}

button.webapp-edit {
  min-width: 36px;
  min-height: 36px;
  padding: 4px;
}

button.webapp-edit image {
  -gtk-icon-size: 22px;
}

button.webapp-x {
  border-radius: 0;
  background: none;
  box-shadow: none;
  border: none;
  padding: 0 6px;
  min-width: 36px;
  min-height: 36px;
  color: #c9e79a;
  font-size: 22px;
  font-weight: 600;
}

button.circular {
  min-width: 52px;
  min-height: 52px;
  padding: 0;
  border-radius: 999px;
  background-color: #28334a;
  color: #d7dde2;
}

button.circular image {
  -gtk-icon-size: 28px;
}

label.webapp-line {
  border-bottom: 2px solid #3d7ea6;
  padding-bottom: 4px;
  font-size: 18px;
}

label.error {
  color: #ff8a80;
}

popover {
  background-color: #1d283a;
}

popover modelbutton,
popover label,
popover button,
popover button label {
  color: #c9e79a;
}
`

func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, "boki-webapps")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	css := filepath.Join(dir, FileName)
	if _, err := os.Stat(css); err != nil {
		if os.IsNotExist(err) {
			if writeErr := os.WriteFile(css, []byte(DefaultCSS), 0o644); writeErr != nil {
				return "", writeErr
			}
		} else {
			return "", err
		}
	}
	return css, nil
}
