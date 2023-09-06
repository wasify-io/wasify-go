package mdk

import "fmt"

//go:wasmimport wasify log
func log(ArgOffset, ArgOffset)

func Log(format string, a ...any) {
	LogDebug(format, a...)
}

func LogDebug(format string, a ...any) {
	slog(format, 1, a...)
}

func LogInfo(format string, a ...any) {
	slog(format, 2, a...)
}

func LogWarning(format string, a ...any) {
	slog(format, 3, a...)
}

func LogError(format string, a ...any) {
	slog(format, 4, a...)
}

func slog(format string, lvl byte, a ...any) {
	log(Arg(fmt.Sprintf(format, a...)), Arg(lvl))
}
