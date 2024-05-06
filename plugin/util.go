package plugin

import (
	"net"
	"os"
	"path/filepath"
)

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
	confPath := filepath.Join(path, ".docker", "config.json")

	if err := os.MkdirAll(confPath, strictFilePerm); err != nil {
		return err
	}

	err := os.WriteFile(path, []byte(conf), strictFilePerm)
	if err != nil {
		return err
	}

	return nil
}
