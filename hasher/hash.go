package hasher

import (
	"fmt"

	gocid "github.com/ipfs/go-cid"
	"golang.org/x/crypto/sha3"
)

const CHUNK_SIZE = 128 * 1024 //128KiB

func HashContent(content []byte) (cidStr string, hash []byte, hashes [][]byte, err error) {
	hashes = make([][]byte, 0)
	merkleHasher := MerkleTreeHash{}

	chunkSize := CHUNK_SIZE
	for i := 0; i < len(content); i += chunkSize {
		end := i + chunkSize
		if end > len(content) {
			end = len(content)
		}
		hashVal := Hash(content[i:end])
		hashes = append(hashes, hashVal)
		merkleHasher.Write(hashVal)
	}

	hash = merkleHasher.Sum(nil)

	b := calBytesFromLength(uint64(len(content)))
	hash = append(hash, b...)
	prefix := gocid.Prefix{
		Version:  1,                           // Usually '1'.
		Codec:    0x55,                        // 0x55 means "raw binary"
		MhType:   META_STORE_MERKLE_TREE_HASH, // use merkle tree to generate data hash
		MhLength: 32 + len(b),                 // pad file size
	}
	cid, err := gocid.Parse(append(prefix.Bytes(), hash...))
	if err != nil {
		return
	}
	cidStr = cid.String()
	return
}

func Hash(content []byte) []byte {
	h := sha3.NewLegacyKeccak256()
	h.Write(content)
	return h.Sum(nil)
}

func GetCidFromHashes(hashes [][]byte, size int) (string, error) {
	hl := len(hashes)
	// 0 < size && 0 < len(hashes)
	if !(0 < size) {
		return "", fmt.Errorf("size[%d] must be greater than 0", size)
	}
	if !(0 < hl) {
		return "", fmt.Errorf("len(hashes)[%d] must be greater than 0", len(hashes))
	}
	// (len(hashes) - 1)*CHUNK_SIZE < size && size <= len(hashes)*CHUNK_SIZE
	if !((hl-1)*CHUNK_SIZE < size && size <= hl*CHUNK_SIZE) {
		return "", fmt.Errorf("len(hashes)[%d] and size[%d] not match", len(hashes), size)
	}

	for i := 0; i < len(hashes); i++ {
		hashVal := hashes[i]
		if len(hashVal) != 32 {
			return "", fmt.Errorf("hashes[%d] length must be 32", len(hashVal))
		}
	}

	merkleHasher := MerkleTreeHash{}
	for i := 0; i < len(hashes); i++ {
		hashVal := hashes[i]
		merkleHasher.Write(hashVal)
	}

	hash := merkleHasher.Sum(nil)

	b := calBytesFromLength(uint64(size))

	hash = append(hash, b...)
	prefix := gocid.Prefix{
		Version:  1,                           // Usually '1'.
		Codec:    0x55,                        // 0x55 means "raw binary"
		MhType:   META_STORE_MERKLE_TREE_HASH, // use merkle tree to generate data hash
		MhLength: 32 + len(b),                 // pad file size
	}
	cid, err := gocid.Parse(append(prefix.Bytes(), hash...))
	if err != nil {
		return "", err
	}
	cidStr := cid.String()
	return cidStr, nil
}

func GetDataLengthFromCid(cid []byte) (uint64, error) {
	prefix, err := gocid.PrefixFromBytes(cid)
	if err != nil {
		return 0, err
	}

	preLen := len(prefix.Bytes())

	sizeBytes := cid[preLen+32:]

	return calLengthFromBytes(sizeBytes), nil
}

func calBytesFromLength(l uint64) []byte {
	size := calLength(l)
	bytes := make([]byte, size)
	for i := 0; i < size; i++ {
		bytes[i] = byte(l >> (8 * i))
	}
	return bytes
}

func calLengthFromBytes(b []byte) uint64 {
	size := len(b)
	l := uint64(0)
	for i := 0; i < size; i++ {
		l += uint64(b[i]) << (8 * i)
	}
	return l
}

func calLength(l uint64) int {
	b := 0
	for ; l > 0; l = (l >> 8) {
		b++
	}

	return b
}
