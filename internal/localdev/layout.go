package localdev

import (
	"fmt"
	"os"
	"path/filepath"
)

const defaultStateDirName = ".opencenter-local"

// Layout describes the local-dev plugin state directory.
type Layout struct {
	Root           string
	GiteaDataDir   string
	GiteaConfDir   string
	GiteaCertDir   string
	TokensDir      string
	MetadataPath   string
	AppIniPath     string
	CACertPath     string
	ServerCertPath string
	ServerKeyPath  string
	AdminTokenPath string
	UserTokenPath  string
}

// ResolveLayout returns the absolute state layout for the plugin.
func ResolveLayout(root string) (Layout, error) {
	if root == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return Layout{}, fmt.Errorf("get current directory: %w", err)
		}
		root = filepath.Join(cwd, defaultStateDirName)
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return Layout{}, fmt.Errorf("resolve state dir: %w", err)
	}

	giteaRoot := filepath.Join(absRoot, "gitea")
	return Layout{
		Root:           absRoot,
		GiteaDataDir:   giteaRoot,
		GiteaConfDir:   filepath.Join(giteaRoot, "gitea", "conf"),
		GiteaCertDir:   filepath.Join(giteaRoot, "gitea", "certs"),
		TokensDir:      filepath.Join(absRoot, "tokens"),
		MetadataPath:   filepath.Join(absRoot, "gitea.json"),
		AppIniPath:     filepath.Join(giteaRoot, "gitea", "conf", "app.ini"),
		CACertPath:     filepath.Join(giteaRoot, "gitea", "certs", "ca.pem"),
		ServerCertPath: filepath.Join(giteaRoot, "gitea", "certs", "cert.pem"),
		ServerKeyPath:  filepath.Join(giteaRoot, "gitea", "certs", "key.pem"),
		AdminTokenPath: filepath.Join(absRoot, "tokens", "gitea-admin.token"),
		UserTokenPath:  filepath.Join(absRoot, "tokens", "gitea-user.token"),
	}, nil
}

// Ensure creates the state directory structure.
func (l Layout) Ensure() error {
	for _, dir := range []string{l.Root, l.GiteaConfDir, l.GiteaCertDir, l.TokensDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create %s: %w", dir, err)
		}
	}
	return nil
}
