package hasher

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
	"io"

	"github.com/cbergoon/merkletree"
	"github.com/multiformats/go-multihash"
)

// This file defines a multihash hasher, which split a byte slice by 32 bytes and then build a merkle tree on them.

const META_STORE_MERKLE_TREE_HASH uint64 = 600 + 1
const META_MERKLE_TREE_HASH_BLOCK_SIZE = 128

// HashedBytes implements merkletree.Content interface only return itself as hashed result
type HashedBytes []byte

func (c HashedBytes) CalculateHash() ([]byte, error) {
	return c, nil
}

func (c HashedBytes) Equals(other merkletree.Content) (bool, error) {
	otherTC, ok := other.(HashedBytes)
	if !ok {
		return false, errors.New("value is not of type HashedBytes")
	}
	return bytes.Equal(otherTC, c), nil
}

type ContentBlock []byte

func (c ContentBlock) CalculateHash() ([]byte, error) {
	h := sha256.New()
	if _, err := h.Write(c); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

func (c ContentBlock) Equals(other merkletree.Content) (bool, error) {
	otherTC, ok := other.(HashedBytes)
	if !ok {
		return false, errors.New("value is not of type ContentBlock")
	}
	return bytes.Equal(otherTC, c), nil
}

type MerkleTreeHash struct {
	buf  bytes.Buffer
	tree *merkletree.MerkleTree
}

func (h *MerkleTreeHash) Write(b []byte) (int, error) {
	return h.buf.Write(b)
}

func (h *MerkleTreeHash) Sum(b []byte) []byte {
	hashList := make([]merkletree.Content, 0)
	for {
		hashBuf := make([]byte, 32)
		n, err := h.buf.Read(hashBuf)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("error reading from reader, err:", err)
			return nil
		}
		if n != 32 {
			hashBuf = append(hashBuf, make([]byte, 32-n)...)
		}
		hashList = append(hashList, HashedBytes(hashBuf))
	}
	t, err := merkletree.NewTree(hashList)
	if err != nil {
		fmt.Println("init merkle tree failed, err:", err)
		return nil
	}
	mr := t.MerkleRoot()
	b = append(b, mr...)
	h.tree = t
	return b
}

func (h *MerkleTreeHash) Reset() {
	h.buf.Reset()
}

func (h *MerkleTreeHash) Size() int {
	return META_MERKLE_TREE_HASH_BLOCK_SIZE
}

func (h *MerkleTreeHash) BlockSize() int {
	return META_MERKLE_TREE_HASH_BLOCK_SIZE
}

func (h *MerkleTreeHash) Marshal() ([]byte, error) {
	if h.tree == nil {
		return nil, errors.New("tree not built")
	}
	return nil, nil
}

func (h *MerkleTreeHash) Unmarshal([]byte) error {

	return nil
}

func init() {
	multihash.Register(META_STORE_MERKLE_TREE_HASH, func() hash.Hash {
		return &MerkleTreeHash{}
	})
}
