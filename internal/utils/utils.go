package utils

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

// calculateHash computes the SHA-256 hash of the input byte slice.
// It returns the hash as a hex-encoded string.
func CalculateHash(data []byte) (hash string, err error) {
	hasher := sha256.New()
	_, err = hasher.Write(data)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// The `ConstantTimeCompare` function is used here to securely compare two hash values.
// It prevents timing-based attacks by ensuring that the comparison takes the same
// amount of time, regardless of whether the values match or not.
// If the hashes are not equal, an error is returned.
func CompareHashes(hash1, hash2 string) error {
	if subtle.ConstantTimeCompare([]byte(hash1), []byte(hash2)) != 1 {
		return fmt.Errorf("the hashes are not equal. needed %s, actual %s", hash1, hash2)
	}

	return nil
}

// uint64ArrayToBytes converts a slice of uint64 integers to a slice of bytes.
// This function is typically used to convert a slice of packed data into bytes,
// which can then be stored in linear memory.
func Uint64ArrayToBytes(data []uint64) []byte {
	// Calculate the total number of bytes required to represent all the uint64
	// integers in the data slice. Since each uint64 integer is 8 bytes long,
	// we multiply the number of uint64 integers by 8 to get the total number of bytes.
	size := len(data) * 8

	result := make([]byte, size)

	for i, d := range data {
		// Convert d to its little-endian byte representation and store it in the
		// result slice. The binary.LittleEndian.PutUint64 function takes a slice
		// of bytes and a uint64 integer, and writes the uint64 integer into the slice
		// of bytes in little-endian order.
		// The result[i<<3:] slice expression ensures that each uint64 integer is
		// written to the correct position in the result slice.
		// i<<3 is equivalent to i*8, but using bit shifting (<<3) is slightly more
		// efficient than multiplication.
		binary.LittleEndian.PutUint64(result[i<<3:], d)
	}

	// Return the result slice of bytes.
	return result
}
