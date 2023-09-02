package wasify

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
	"github.com/puzpuzpuz/xsync/v2"
)

type LogSeverity uint8

// The log level is initially set to "Info" for runtimes and "zero" (0) for modules.
// However, modules will adopt the log level from their parent runtime.
// If you want only "Error" level for a runtime but need to debug specific module(s),
// you can set those modules to "Debug". This will replace the inherited log level,
// allowing the module to display debug information.
const (
	LogDebug LogSeverity = iota + 1
	LogInfo
	LogWarning
	LogError
)

var logMap = map[LogSeverity]slog.Level{
	LogDebug:   slog.LevelDebug,
	LogInfo:    slog.LevelInfo,
	LogWarning: slog.LevelWarn,
	LogError:   slog.LevelError,
}

// newLogger returns new slog ref
func newLogger(severity LogSeverity) *slog.Logger {

	w := os.Stderr
	logger := slog.New(tint.NewHandler(w, &tint.Options{
		Level:      getlogLevel(severity),
		TimeFormat: time.Kitchen,
		NoColor:    !isatty.IsTerminal(w.Fd()),
	}))

	return logger
}

// getlogLevel gets 'slog' level based on severity specified by user
func getlogLevel(s LogSeverity) slog.Level {

	val, ok := logMap[s]
	if !ok {
		// default logger is Info
		return logMap[LogInfo]
	}

	return val
}

// calculateHash computes the SHA-256 hash of the input byte slice.
// It returns the hash as a hex-encoded string.
func calculateHash(data []byte) (hash string, err error) {
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
func compareHashes(hash1, hash2 string) error {
	if subtle.ConstantTimeCompare([]byte(hash1), []byte(hash2)) != 1 {
		return fmt.Errorf("the hashes are not equal. needed %s, actual %s", hash1, hash2)
	}

	return nil
}

// uint64ArrayToBytes converts a slice of uint64 integers to a slice of bytes.
// This function is typically used to convert a slice of packed data into bytes,
// which can then be stored in linear memory.
func uint64ArrayToBytes(data []uint64) []byte {
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

// This is a custom map-like structure designed for managing allocations.
// It keeps track of offset and size values, and provides methods
// for storing, loading, deleting, and calculating the total size and etc.
//
// allocationMap is employed to monitor allocations made for parameters and return values
// within host functions. These allocations can be automatically cleared later,
// relieving users from the need to manually manage them.
type allocationMap[K xsync.IntegerConstraint, V xsync.IntegerConstraint] struct {
	_map  *xsync.MapOf[K, V]
	_size V
}

func newAllocationMap[K xsync.IntegerConstraint, V xsync.IntegerConstraint]() *allocationMap[K, V] {
	return &allocationMap[K, V]{
		_map: xsync.NewIntegerMapOf[K, V](),
	}
}

func (am *allocationMap[K, V]) store(offset K, size V) {
	am._map.Store(offset, size)
	am._size += size
}

func (am *allocationMap[K, V]) load(offset K) (V, bool) {
	return am._map.Load(offset)
}

func (am *allocationMap[K, V]) delete(offset K) {
	v, _ := am._map.LoadAndDelete(offset)
	am._size -= v
}

func (am *allocationMap[K, V]) totalSize() V {
	return am._size
}
