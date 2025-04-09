package plugin

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
)

var errInvalidDockerConfig = fmt.Errorf("invalid docker config")

func GetContainerIP() (string, error) {
	netInterfaceAddrList, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, netInterfaceAddr := range netInterfaceAddrList {
		netIP, ok := netInterfaceAddr.(*net.IPNet)
		if ok && !netIP.IP.IsLoopback() && netIP.IP.To4() != nil {
			return netIP.IP.String(), nil
		}
	}

	return "", nil
}

func WriteDockerConf(path, conf string) error {
	var jsonData interface{}
	if err := json.Unmarshal([]byte(conf), &jsonData); err != nil {
		return fmt.Errorf("%w: %w", errInvalidDockerConfig, err)
	}

	jsonBytes, err := json.Marshal(jsonData)
	if err != nil {
		return fmt.Errorf("%w: %w", errInvalidDockerConfig, err)
	}

	err = os.WriteFile(path, jsonBytes, strictFilePerm)
	if err != nil {
		return err
	}

	return nil
}

func (p *Plugin) GenerateLabels() []string {
	l := make([]string, 0)

	// As described in https://github.com/opencontainers/image-spec/blob/main/annotations.md
	l = append(l, fmt.Sprintf("org.opencontainers.image.created=%s", p.Settings.Build.Time))

	if p.Settings != nil {
		if tags := p.Settings.Build.Tags; len(tags) > 0 {
			l = append(l, fmt.Sprintf("org.opencontainers.image.version=%s", tags[len(tags)-1]))
		}
	}

	if p.Repository != nil && p.Repository.URL != "" {
		l = append(l, fmt.Sprintf("org.opencontainers.image.source=%s", p.Repository.URL))
		l = append(l, fmt.Sprintf("org.opencontainers.image.url=%s", p.Repository.URL))
	}

	if p.Commit != nil && p.Commit.SHA != "" {
		l = append(l, fmt.Sprintf("org.opencontainers.image.revision=%s", p.Commit.SHA))
	}

	return l
}
