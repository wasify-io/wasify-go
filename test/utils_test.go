package test

import (
	"log/slog"
	"testing"

	"github.com/wasify-io/wasify-go/internal/memory"
	"github.com/wasify-io/wasify-go/internal/utils"
)

func TestGetLogLevel(t *testing.T) {
	tests := []struct {
		severity utils.LogSeverity
		expected slog.Level
	}{
		{utils.LogDebug, slog.LevelDebug},
		{utils.LogInfo, slog.LevelInfo},
		{utils.LogWarning, slog.LevelWarn},
		{utils.LogError, slog.LevelError},
		{utils.LogSeverity(255), slog.LevelInfo}, // Unexpected severity
	}

	for _, test := range tests {
		got := utils.GetlogLevel(test.severity)
		if got != test.expected {
			t.Errorf("for severity %d, expected %d but got %d", test.severity, test.expected, got)
		}
	}
}

func TestCalculateHash(t *testing.T) {
	data := []byte("test")
	expected := "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08" // known hash for "test"

	hash, err := utils.CalculateHash(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash != expected {
		t.Errorf("expected hash %s, but got %s", expected, hash)
	}
}

func TestCompareHashes(t *testing.T) {
	hash := "wrong hash"

	err := utils.CompareHashes(hash, hash)
	if err != nil {
		t.Errorf("did not expect an error for equal hashes, but got %v", err)
	}

	err = utils.CompareHashes(hash, "12345")
	if err == nil {
		t.Error("expected an error for different hashes, but got none")
	}
}

func TestUint64ArrayToBytes(t *testing.T) {
	data := []uint64{1, 2}
	expected := []byte{1, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0} // little-endian representation

	result := utils.Uint64ArrayToBytes(data)
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
	am := memory.NewAllocationMap[int, int]()
	am.Store(1, 100)
	am.Store(2, 200)

	if size, _ := am.Load(1); size != 100 {
		t.Errorf("expected size 100 for offset 1, but got %d", size)
	}

	am.Delete(1)
	if _, exists := am.Load(1); exists {
		t.Error("expected offset 1 to be deleted, but it still exists")
	}

	if total := am.TotalSize(); total != 200 {
		t.Errorf("expected total size to be 200, but got %d", total)
	}
}
