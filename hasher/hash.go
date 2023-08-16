package hasher

import (
	"fmt"

	merkletree "github.com/MetaDataLab/go-merkletree"
	gocid "github.com/ipfs/go-cid"
)

const CHUNK_SIZE = 128 * 1024 //128KiB
const HashCode = merkletree.KECCAK256
const META_STORE_MERKLE_TREE_HASH uint64 = 600 + 1

func HashContent(content []byte) (cidStr string, hash []byte, hashes [][]byte, err error) {
	data := [][]byte{}

	chunkSize := CHUNK_SIZE
	for i := 0; i < len(content); i += chunkSize {
		end := i + chunkSize
		if end > len(content) {
			end = len(content)
		}
		data = append(data, content[i:end])
	}

	hashType, err := merkletree.GetHashTypeFromCode(HashCode)
	if err != nil {
		return
	}

	tree, err := merkletree.NewTree(
		merkletree.WithData(data),
		merkletree.WithHashType(hashType),
	)
	if err != nil {
		return
	}
	hash = tree.Root()
	hashes = tree.LeavesNodes()

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
	h, _ := merkletree.GetHashTypeFromCode(HashCode)
	return h.Hash(content)
}

func Hashes(content []byte) (hashes [][]byte) {
	hashType, err := merkletree.GetHashTypeFromCode(HashCode)
	if err != nil {
		return
	}

	for i := 0; i < len(content); i += CHUNK_SIZE {
		end := i + CHUNK_SIZE
		if end > len(content) {
			end = len(content)
		}
		hashVal := hashType.Hash(content[i:end])
		hashes = append(hashes, hashVal)
	}

	return
}

func HashHashes(hashes [][]byte, size int) (cidStr string, hash []byte, err error) {
	hl := len(hashes)
	// 0 < size && 0 < len(hashes)
	if !(0 < size) {
		return "", nil, fmt.Errorf("size[%d] must be greater than 0", size)
	}
	if !(0 < hl) {
		return "", nil, fmt.Errorf("len(hashes)[%d] must be greater than 0", len(hashes))
	}
	if !((hl-1)*CHUNK_SIZE < size && size <= hl*CHUNK_SIZE) {
		return "", nil, fmt.Errorf("len(hashes)[%d] and size[%d] not match", len(hashes), size)
	}

	for i := 0; i < len(hashes); i++ {
		hashVal := hashes[i]
		if len(hashVal) != 32 {
			return "", nil, fmt.Errorf("hashes[%d] length must be 32", len(hashVal))
		}
	}

	h, _ := merkletree.GetHashTypeFromCode(HashCode)

	tree, err := merkletree.NewTreeWithLeavesHashes(hashes, h)
	if err != nil {
		return "", nil, err
	}
	hash = tree.Root()

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
		return "", nil, err
	}
	cidStr = cid.String()
	return
}

// deprecated, use HashHashes instead
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

	h, _ := merkletree.GetHashTypeFromCode(HashCode)

	tree, err := merkletree.NewTreeWithLeavesHashes(hashes, h)
	if err != nil {
		return "", err
	}
	hash := tree.Root()

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
