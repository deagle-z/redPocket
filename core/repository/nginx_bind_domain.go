package repository

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

const (
	nginxConfDir = "/etc/nginx/conf.d"
	h5RootDir    = "/home/service/red/h5"
)

var bindDomainPattern = regexp.MustCompile(`^(\*\.)?([a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,63}$`)

func normalizeBindDomain(bindDomain *string) (*string, error) {
	if bindDomain == nil {
		return nil, nil
	}

	domain := strings.ToLower(strings.TrimSpace(*bindDomain))
	if domain == "" {
		return nil, nil
	}
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimRight(domain, "/")
	if !bindDomainPattern.MatchString(domain) {
		return nil, errors.New("invalid_bind_domain")
	}
	return &domain, nil
}

func createNginxBindDomainConfig(bindDomain string) error {
	if runtime.GOOS != "linux" {
		return nil
	}

	if err := os.MkdirAll(nginxConfDir, 0755); err != nil {
		return fmt.Errorf("create_nginx_conf_dir_failed: %w", err)
	}

	confPath := nginxBindDomainConfPath(bindDomain)
	file, err := os.OpenFile(confPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		if os.IsExist(err) {
			return errors.New("bind_domain_nginx_conf_exists")
		}
		return fmt.Errorf("create_nginx_conf_failed: %w", err)
	}

	_, writeErr := file.WriteString(buildNginxBindDomainConfig(bindDomain))
	closeErr := file.Close()
	if writeErr != nil {
		_ = os.Remove(confPath)
		return fmt.Errorf("write_nginx_conf_failed: %w", writeErr)
	}
	if closeErr != nil {
		_ = os.Remove(confPath)
		return fmt.Errorf("close_nginx_conf_failed: %w", closeErr)
	}

	if output, err := exec.Command("nginx", "-t").CombinedOutput(); err != nil {
		_ = os.Remove(confPath)
		return fmt.Errorf("nginx_test_failed: %s", strings.TrimSpace(string(output)))
	}
	if output, err := exec.Command("nginx", "-s", "reload").CombinedOutput(); err != nil {
		_ = os.Remove(confPath)
		_, _ = exec.Command("nginx", "-t").CombinedOutput()
		_, _ = exec.Command("nginx", "-s", "reload").CombinedOutput()
		return fmt.Errorf("nginx_reload_failed: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

func nginxBindDomainConfPath(bindDomain string) string {
	fileName := strings.ReplaceAll(bindDomain, "*.", "wildcard.")
	fileName = strings.ReplaceAll(fileName, ".", "_")
	return filepath.Join(nginxConfDir, fileName+".conf")
}

func buildNginxBindDomainConfig(bindDomain string) string {
	return fmt.Sprintf(`server {
    listen 80;
    server_name %s;

    root %s;
    index index.html index.htm;

    location / {
        try_files $uri $uri/ /index.html;
    }
}
`, bindDomain, h5RootDir)
}
