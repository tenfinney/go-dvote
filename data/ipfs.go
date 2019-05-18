package data

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os/exec"
	"time"

	"github.com/vocdoni/go-dvote/types"

	shell "github.com/ipfs/go-ipfs-api"
)

type IPFSConfig struct {
	Binary      string
	InitTimeout int
}

type IPFSHandle struct {
	d *types.DataStore
	s *shell.Shell
	c *IPFSConfig
	//can we add a shell here for use by all methods?
}

func IPFSNewConfig() *IPFSConfig {
	cfg := new(IPFSConfig)
	cfg.Binary = "/usr/local/bin/ipfs"
	cfg.InitTimeout = 10
	return cfg
}

// init IPFS daemon with pubsub features enabled
func startIPFSDaemon(ipfsBinPath string) (err error) {
	cmd := exec.Command(ipfsBinPath, "daemon")
	if err := cmd.Start(); err != nil {
		return errors.New("Cannot init the IPFS daemon")
	}
	return nil
}

func (i *IPFSHandle) Init(d *types.DataStore) error {
	err := startIPFSDaemon(i.c.Binary)
	if err != nil {
		return err
	}
	i.d = d
	i.s = shell.NewShell("localhost:5001")
	for timeout := i.c.InitTimeout; timeout > 0; timeout-- {
		if i.s.IsUp() {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	//test that ipfs is running/working
	return errors.New("Could not connect to IPFS daemon")
}

func (i *IPFSHandle) Publish(object []byte) (string, error) {
	cid, err := i.s.Add(bytes.NewBuffer(object))
	if err != nil {
		return "", err
	}
	return cid, nil
}

func (i *IPFSHandle) Pin(path string) error {
	sh := shell.NewShell("localhost:5001")
	err := sh.Pin(path)
	if err != nil {
		return err
	}
	return nil
}

func (i *IPFSHandle) Unpin(path string) error {
	sh := shell.NewShell("localhost:5001")
	err := sh.Unpin(path)
	if err != nil {
		return err
	}
	return nil
}

func (i *IPFSHandle) ListPins() (map[string]string, error) {
	sh := shell.NewShell("localhost:5001")
	info, err := sh.Pins()
	if err != nil {
		return nil, err
	}
	var pinMap map[string]string
	pinMap = make(map[string]string)
	for k, c := range info {
		pinMap[k] = c.Type
	}
	return pinMap, nil
}

func (i *IPFSHandle) Retrieve(hash string) ([]byte, error) {
	sh := shell.NewShell("localhost:5001")
	reader, err := sh.Cat(hash)
	if err != nil {
		return nil, err
	}
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return content, nil
}
