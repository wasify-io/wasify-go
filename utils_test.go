package wasify

import (
	"log/slog"
	"testing"
)

func TestGetLogLevel(t *testing.T) {
	tests := []struct {
		severity LogSeverity
		expected slog.Level
	}{
		{LogDebug, slog.LevelDebug},
		{LogInfo, slog.LevelInfo},
		{LogWarning, slog.LevelWarn},
		{LogError, slog.LevelError},
		{LogSeverity(255), slog.LevelInfo}, // Unexpected severity
	}

	for _, test := range tests {
		got := getlogLevel(test.severity)
		if got != test.expected {
			t.Errorf("for severity %d, expected %d but got %d", test.severity, test.expected, got)
		}
	}
}

func TestCalculateHash(t *testing.T) {
	data := []byte("test")
	expected := "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08" // known hash for "test"

	hash, err := calculateHash(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash != expected {
		t.Errorf("expected hash %s, but got %s", expected, hash)
	}
}

func TestCompareHashes(t *testing.T) {
	hash := "wrong hash"

	err := compareHashes(hash, hash)
	if err != nil {
		t.Errorf("did not expect an error for equal hashes, but got %v", err)
	}

	err = compareHashes(hash, "12345")
	if err == nil {
		t.Error("expected an error for different hashes, but got none")
	}
}

func TestUint64ArrayToBytes(t *testing.T) {
	data := []uint64{1, 2}
	expected := []byte{1, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0} // little-endian representation

	result := uint64ArrayToBytes(data)
	if len(result) != len(expected) {
		t.Fatalf("expected %d bytes but got %d", len(expected), len(result))
	}

	for i, b := range result {
		if b != expected[i] {
			t.Errorf("at index %d: expected byte %d but got %d", i, expected[i], b)
		}
	}
}

func TestAllocationMap(t *testing.T) {
	am := newAllocationMap[int, int]()
	am.store(1, 100)
	am.store(2, 200)

	if size, _ := am.load(1); size != 100 {
		t.Errorf("expected size 100 for offset 1, but got %d", size)
	}

	am.delete(1)
	if _, exists := am.load(1); exists {
		t.Error("expected offset 1 to be deleted, but it still exists")
	}

	if total := am.totalSize(); total != 200 {
		t.Errorf("expected total size to be 200, but got %d", total)
	}
}
