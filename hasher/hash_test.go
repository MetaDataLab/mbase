package hasher

import (
	"bytes"
	"testing"
)

func TestGetCidFromHashes(t *testing.T) {
	size := 127*1024 + 1025
	cid := "bafk5sbbdv6ff5evom74oaepgwvvmskovoarzere35imcurv4f2bo5vnk5sqqcaac"

	hash0 := []byte{51, 122, 33, 203, 126, 184, 158, 214, 27, 236, 117, 181, 214, 250, 209, 27, 68, 104, 132, 173, 64, 186, 164, 66, 96, 5, 63, 219, 191, 23, 246, 58}
	hash1 := []byte{181, 85, 61, 227, 21, 224, 237, 245, 4, 217, 21, 10, 248, 45, 175, 165, 196, 102, 127, 166, 24, 237, 10, 111, 25, 198, 155, 65, 22, 108, 85, 16}
	hashes := make([][]byte, 0)
	hashes = append(hashes, hash0)
	hashes = append(hashes, hash1)

	c, err := GetCidFromHashes(hashes, size)
	if err != nil {
		t.Errorf("GetCidFromHashes error: %s", err.Error())
	}

	if cid != c {
		t.Errorf("cid not match")
	}
}

func TestHashContent(t *testing.T) {

	data := make([]byte, 0)
	for i := 0; i < 127*1024; i++ {
		data = append(data, 'a')
	}
	for i := 0; i < 1025; i++ {
		data = append(data, 'b')
	}

	cid := "bafk5sbbdv6ff5evom74oaepgwvvmskovoarzere35imcurv4f2bo5vnk5sqqcaac"
	rootHash := []byte{175, 138, 94, 146, 174, 103, 248, 224, 17, 230, 181, 106, 201, 41, 213, 112, 35, 146, 68, 155, 234, 24, 42, 70, 188, 46, 130, 238, 213, 170, 236, 161, 1, 0, 2}
	hash0 := []byte{51, 122, 33, 203, 126, 184, 158, 214, 27, 236, 117, 181, 214, 250, 209, 27, 68, 104, 132, 173, 64, 186, 164, 66, 96, 5, 63, 219, 191, 23, 246, 58}
	hash1 := []byte{181, 85, 61, 227, 21, 224, 237, 245, 4, 217, 21, 10, 248, 45, 175, 165, 196, 102, 127, 166, 24, 237, 10, 111, 25, 198, 155, 65, 22, 108, 85, 16}
	hashes := make([][]byte, 0)
	hashes = append(hashes, hash0)
	hashes = append(hashes, hash1)

	c, r, hs, _ := HashContent(data)
	if c != cid {
		t.Errorf("cid not match, cid %s, c %s", cid, c)
	}

	if !bytes.Equal(rootHash, r) {
		t.Errorf("rootHash and not the same, rootHash %v, r %v", rootHash, r)
	}

	if len(hs) != len(hashes) {
		t.Errorf("hashes length not match, len(hs) %d, len(hashes) %d", len(hs), len(hashes))
	}

	for i := 0; i < len(hashes); i++ {
		if !bytes.Equal(hashes[i], hs[i]) {
			t.Errorf("hashes[%d] %s != hs[%d] %s", i, hashes[i], i, hs[i])
		}
	}
}

func TestGetDataLengthFromCid(t *testing.T) {
	cidBytes := []byte{1, 85, 217, 4, 35, 105, 219, 140, 113, 22, 200, 38, 34, 214, 59, 196, 200, 49, 213, 55, 215, 167, 97, 24, 18, 192, 67, 180, 239, 245, 190, 209, 35, 190, 4, 132, 134, 1, 0, 2}
	size := 127*1024 + 1025
	l, err := GetDataLengthFromCid(cidBytes)
	if err != nil {
		t.Errorf("GetDataLengthFromCid failed, %s", err.Error())
	}

	if l != uint64(size) {
		t.Errorf("l %d != size %d", l, size)
	}
}

func TestHashes(t *testing.T) {

	data := make([]byte, 0)
	for i := 0; i < 127*1024; i++ {
		data = append(data, 'a')
	}
	for i := 0; i < 1025; i++ {
		data = append(data, 'b')
	}

	hash0 := []byte{51, 122, 33, 203, 126, 184, 158, 214, 27, 236, 117, 181, 214, 250, 209, 27, 68, 104, 132, 173, 64, 186, 164, 66, 96, 5, 63, 219, 191, 23, 246, 58}
	hash1 := []byte{181, 85, 61, 227, 21, 224, 237, 245, 4, 217, 21, 10, 248, 45, 175, 165, 196, 102, 127, 166, 24, 237, 10, 111, 25, 198, 155, 65, 22, 108, 85, 16}
	hashes := make([][]byte, 0)
	hashes = append(hashes, hash0)
	hashes = append(hashes, hash1)

	hs := Hashes(data)

	for i := 0; i < len(hashes); i++ {
		if !bytes.Equal(hashes[i], hs[i]) {
			t.Errorf("hashes[%d] %s != hs[%d] %s", i, hashes[i], i, hs[i])
		}
	}
}
