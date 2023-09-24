package mdk

import "fmt"

//go:wasmimport wasify log
func _log(PackedData, PackedData)

func Log(format string, a ...any) {
	LogDebug(format, a...)
}

func LogDebug(format string, a ...any) {
	_slog(format, 1, a...)
}

func LogInfo(format string, a ...any) {
	_slog(format, 2, a...)
}

func LogWarning(format string, a ...any) {
	_slog(format, 3, a...)
}

func LogError(format string, a ...any) {
	_slog(format, 4, a...)
}

func _slog(format string, lvl byte, a ...any) {
	_log(Arg(fmt.Sprintf(format, a...)), Arg(lvl))
}
