package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ExpandPath expands ~ and environment variables in a path.
func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			path = filepath.Join(home, path[2:])
		}
	}
	return os.ExpandEnv(path)
}

// OpenWithDefaultApp opens a file path or URL with the system default handler.
func OpenWithDefaultApp(path string) error {
	path = ExpandPath(path)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", path)
	default:
		for _, opener := range []string{"xdg-open", "gio", "gnome-open", "kde-open"} {
			if _, err := exec.LookPath(opener); err == nil {
				cmd = exec.Command(opener, path)
				break
			}
		}
		if cmd == nil {
			return fmt.Errorf("no suitable opener found for your system")
		}
	}

	return cmd.Start()
}
