package data

import (
	"errors"

	"github.com/vocdoni/go-dvote/types"
)

type Storage interface {
	Init(d *types.DataStore) error
	Publish(o []byte) (string, error)
	Retrieve(id string) ([]byte, error)
	Pin(path string) error
	Unpin(path string) error
	ListPins() (map[string]string, error)
}

type StorageConfig interface {
	Type() StorageID
}

type StorageID int

const (
	IPFS StorageID = iota + 1
	BZZ
)

func StorageIDFromString(i string) StorageID {
	switch i {
	case "IPFS":
		return IPFS
	case "BZZ":
		return BZZ
	default:
		return -1
	}
}

func InitDefault(t StorageID, config *StorageConfig) (Storage, error) {
	switch t {
	case IPFS:
		s := new(IPFSHandle)
		//s.c = IPFSNewConfig()
		s.c = config.(*IPFSConfig)
		defaultDataStore := new(types.DataStore)
		defaultDataStore.Datadir = "this_is_still_ignored"
		err := s.Init(defaultDataStore)
		return s, err
	case BZZ:
		s := new(BZZHandle)
		defaultDataStore := new(types.DataStore)
		defaultDataStore.Datadir = "this_is_still_ignored"
		err := s.Init(defaultDataStore)
		return s, err
	default:
		return nil, errors.New("Bad storage type specification")
	}
}

func Init(t StorageID, d *types.DataStore) (Storage, error) {
	switch t {
	case IPFS:
		s := new(IPFSHandle)
		s.Init(d)
		return s, nil
	case BZZ:
		s := new(BZZHandle)
		s.Init(d)
		return s, nil
	default:
		return nil, errors.New("Bad storage type or DataStore specification")
	}
}
