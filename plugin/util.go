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
