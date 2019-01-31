package tree

import (
	"errors"
	"fmt"
	"os/user"

	common3 "github.com/iden3/go-iden3/common"
	mkcore "github.com/iden3/go-iden3/core"
	db "github.com/iden3/go-iden3/db"
	merkletree "github.com/iden3/go-iden3/merkletree"
)

type Tree struct {
	Namespace string
	Storage   string
	Tree      *merkletree.MerkleTree
	DbStorage *db.LevelDbStorage
}

func (t *Tree) Init() error {
	if len(t.Storage) < 1 {
		usr, err := user.Current()
		if err == nil {
			t.Storage = usr.HomeDir + "/.dvote/Tree"
		} else {
			t.Storage = "./dvoteTree"
		}
	}
	mtdb, err := db.NewLevelDbStorage(t.Storage, false)
	if err != nil {
		return err
	}
	mt, err := merkletree.NewMerkleTree(mtdb, 140)
	if err != nil {
		return err
	}
	t.DbStorage = mtdb
	t.Tree = mt
	return nil
}

func (t *Tree) Close() {
	defer t.Tree.Storage().Close()
}

func (t *Tree) GetClaim(data []byte) (*mkcore.ClaimBasic, error) {
	if len(data) > 496/8 {
		return nil, errors.New("claim data too large")
	}
	for i := len(data); i <= 496/8; i++ {
		data = append(data, byte('0'))
	}
	var indexSlot [400 / 8]byte
	var dataSlot [496 / 8]byte
	copy(indexSlot[:], data[:400/8])
	copy(dataSlot[:], data[:496/8])
	e := mkcore.NewClaimBasic(indexSlot, dataSlot)
	return e, nil
}

func (t *Tree) AddClaim(data []byte) error {
	e, err := t.GetClaim(data)
	if err != nil {
		return err
	}
	return t.Tree.Add(e.Entry())
}

func (t *Tree) GenProof(data []byte) (string, error) {
	e, err := t.GetClaim(data)
	if err != nil {
		return "", err
	}
	mp, err := t.Tree.GenerateProof(e.Entry().HIndex())
	if err != nil {
		return "", err
	}
	mpHex := common3.HexEncode(mp.Bytes())
	return mpHex, nil
}

func (t *Tree) CheckProof(data []byte, mpHex string) (bool, error) {
	mpBytes, err := common3.HexDecode(mpHex)
	if err != nil {
		return false, err
	}
	mp, err := merkletree.NewProofFromBytes(mpBytes)
	if err != nil {
		return false, err
	}
	e, err := t.GetClaim(data)
	if err != nil {
		return false, err
	}
	return merkletree.VerifyProof(t.Tree.RootKey(), mp,
		e.Entry().HIndex(), e.Entry().HValue()), nil
}

func (t *Tree) GetRoot() string {
	return t.Tree.RootKey().String()
}

func (t *Tree) Dump() ([]string, error) {
	var response []string

	err := t.Tree.Walk(t.Tree.RootKey(), func(n *merkletree.Node) {
		if n.Type == merkletree.NodeTypeLeaf {
			response = append(response, fmt.Sprintf("|%s", n.Entry.Data))
		}
	})
	return response, err
}

func (t *Tree) Snapshot() (string, error) {
	snapshotNamespace := t.Tree.RootKey().String()
	return snapshotNamespace, nil
}
