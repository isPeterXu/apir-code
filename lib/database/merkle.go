package database

import (
	"encoding/binary"
	"io"
	"log"

	"github.com/si-co/vpir-code/lib/merkle"
)

// CreateRandomMultiBitMerkle
// blockLen is the number of byte in a block, as byte is viewd as an element in this
// case
func CreateRandomMultiBitMerkle(rnd io.Reader, dbLen, numRows, blockLen int) *Bytes {
	entries := make([][]byte, numRows)
	numBlocks := dbLen / (8 * blockLen)
	// generate random blocks
	blocks := make([][]byte, numBlocks)
	for i := range blocks {
		// generate random block
		b := make([]byte, blockLen)
		if _, err := rnd.Read(b); err != nil {
			log.Fatal(err)
		}
		blocks[i] = b
	}

	// generate tree
	tree, err := merkle.New(blocks)
	if err != nil {
		log.Fatalf("impossible to create Merkle tree: %v", err)
	}

	// generate db
	blocksPerRow := numBlocks / numRows
	proofLen := 0
	b := 0
	for i := range entries {
		e := make([]byte, 0)
		for j := 0; j < blocksPerRow; j++ {
			p, err := tree.GenerateProof(blocks[b])
			encodedProof := encodeProof(p)
			if err != nil {
				log.Fatalf("error while generating proof for block %v: %v", b, err)
			}
			e = append(e, append(blocks[b], encodedProof...)...)
			proofLen = len(encodedProof) // always same length
			b++
		}
		entries[i] = e
	}
	root := tree.Root()

	m := &Bytes{
		Entries: entries,
		Info: Info{
			NumRows:    numRows,
			NumColumns: dbLen / (8 * numRows * blockLen),
			BlockSize:  blockLen + proofLen,
			PIRType:    "merkle",
			Root:       root,
			ProofLen:   proofLen,
		},
	}

	return m
}

func DecodeProof(p []byte) *merkle.Proof {
	// number of hashes
	numHashes := binary.LittleEndian.Uint32(p[0:])

	// hashes
	hashLength := uint32(32) // sha256
	hashes := make([][]byte, numHashes)
	for i := uint32(0); i < numHashes; i++ {
		hashes[i] = p[4+hashLength*i : 4+hashLength*(i+1)]
	}

	// index
	index := binary.LittleEndian.Uint64(p[len(p)-8:])

	return &merkle.Proof{
		Hashes: hashes,
		Index:  index,
	}
}

func encodeProof(p *merkle.Proof) []byte {
	out := make([]byte, 0)

	// encode number of hashes
	numHashes := uint32(len(p.Hashes))
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, numHashes)
	out = append(out, b...)

	// encode hashes
	for _, h := range p.Hashes {
		out = append(out, h...)
	}

	// encode index
	b1 := make([]byte, 8)
	binary.LittleEndian.PutUint64(b1, p.Index)
	out = append(out, b1...)

	return out
}
