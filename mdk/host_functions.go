package mdk

import "fmt"

//go:wasmimport myEnv log
func log(ArgOffset)

func Log(format string, a ...any) {
	log(Arg(fmt.Sprintf(format, a...)))
}
