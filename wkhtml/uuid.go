package wkhtml

import (
	"crypto/rand"
	"encoding/hex"
	"io"
)

// NewUUID creates a new random UUID or panics.
func NewUUID() UUID {
	return MustNewUUID(NewRandomUUID())
}

// MustNewUUID returns uuid if err is nil and panics otherwise.
func MustNewUUID(uuid UUID, err error) UUID {
	if err != nil {
		panic(err)
	}
	return uuid
}

// String returns the string form of uuid, xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
// , or "" if uuid is invalid.
func (uuid UUID) String() string {
	var buf [36]byte
	encodeHex(buf[:], uuid)
	return string(buf[:])
}

func encodeHex(dst []byte, uuid UUID) {
	hex.Encode(dst, uuid[:4])
	dst[8] = '-'
	hex.Encode(dst[9:13], uuid[4:6])
	dst[13] = '-'
	hex.Encode(dst[14:18], uuid[6:8])
	dst[18] = '-'
	hex.Encode(dst[19:23], uuid[8:10])
	dst[23] = '-'
	hex.Encode(dst[24:], uuid[10:])
}

var rander = rand.Reader // random function
var Nil UUID             // empty UUID, all zeros

// A UUID is a 128 bit (16 byte) Universal Unique IDentifier as defined in RFC 4122.
type UUID [16]byte

// NewRandomUUID returns a Random (Version 4) UUID.
//
// The strength of the UUIDs is based on the strength of the crypto/rand
// package.
//
// A note about uniqueness derived from the UUID Wikipedia entry:
//
//  Randomly generated UUIDs have 122 random bits.  One's annual risk of being
//  hit by a meteorite is estimated to be one chance in 17 billion, that
//  means the probability is about 0.00000000006 (6 × 10−11),
//  equivalent to the odds of creating a few tens of trillions of UUIDs in a
//  year and having one duplicate.
func NewRandomUUID() (UUID, error) {
	return NewRandomUUIDFromReader(rander)
}

// NewRandomUUIDFromReader returns a UUID based on bytes read from a given io.Reader.
func NewRandomUUIDFromReader(r io.Reader) (UUID, error) {
	var uuid UUID
	_, err := io.ReadFull(r, uuid[:])
	if err != nil {
		return Nil, err
	}
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant is 10
	return uuid, nil
}
