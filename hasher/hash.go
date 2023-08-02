package hasher

import (
	"unsafe"

	gocid "github.com/ipfs/go-cid"
	"golang.org/x/crypto/sha3"
)

const CHUNK_SIZE = 128 * 1024 //128KiB

func HashContent(content []byte) (cidStr string, hashes [][]byte, err error) {
	hashes = make([][]byte, 0)
	merkleHasher := MerkleTreeHash{}

	chunkSize := CHUNK_SIZE
	for i := 0; i < len(content); i += chunkSize {
		end := i + chunkSize
		if end > len(content) {
			end = len(content)
		}
		h := sha3.NewLegacyKeccak256()
		h.Write(content[i:end])
		hashVal := h.Sum(nil)
		hashes = append(hashes, hashVal)
		merkleHasher.Write(hashVal)
	}

	hash := merkleHasher.Sum(nil)

	b := CalBytesFromLength(uint64(len(content)))
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

func GetDataLengthFromCid(cid []byte) (uint64, error) {
	prefix, err := gocid.PrefixFromBytes(cid)
	if err != nil {
		return 0, err
	}

	preLen := unsafe.Sizeof(prefix)

	sizeBytes := cid[preLen+32:]

	return CalLengthFromBytes(sizeBytes), nil
}

func CalBytesFromLength(l uint64) []byte {
	size := calLength(l)
	bytes := make([]byte, size)
	for i := 0; i < size; i++ {
		bytes[i] = byte(l >> (8 * i))
	}
	return bytes
}

func CalLengthFromBytes(b []byte) uint64 {
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
