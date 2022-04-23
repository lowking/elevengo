package ec115

import (
	"crypto/elliptic"
	"crypto/rand"
	"encoding/binary"
)

var (
	curve = elliptic.P224()

	svrX, svrY = elliptic.Unmarshal(curve, []byte{
		0x04, 0x57, 0xa2, 0x92, 0x57, 0xcd, 0x23, 0x20,
		0xe5, 0xd6, 0xd1, 0x43, 0x32, 0x2f, 0xa4, 0xbb,
		0x8a, 0x3c, 0xf9, 0xd3, 0xcc, 0x62, 0x3e, 0xf5,
		0xed, 0xac, 0x62, 0xb7, 0x67, 0x8a, 0x89, 0xc9,
		0x1a, 0x83, 0xba, 0x80, 0x0d, 0x61, 0x29, 0xf5,
		0x22, 0xd0, 0x34, 0xc8, 0x95, 0xdd, 0x24, 0x65,
		0x24, 0x3a, 0xdd, 0xc2, 0x50, 0x95, 0x3b, 0xee,
		0xba,
	})

	salt = []byte("^j>WD3Kr?J2gLFjD4W2y@")

	le = binary.LittleEndian
)

func (c *Coder) init() *Coder {
	// Generate EC key-pair
	privKey, x, y, _ := elliptic.GenerateKey(curve, rand.Reader)
	pubKey := elliptic.MarshalCompressed(curve, x, y)

	// Store public key
	c.pubKey[0] = 0x1d
	copy(c.pubKey[1:], pubKey)

	// ECDH key exchanging
	x, _ = curve.ScalarMult(svrX, svrY, privKey)
	secret := x.Bytes()
	copy(c.aesKey, secret[:16])
	copy(c.aesIv, secret[len(secret)-16:])

	return c
}