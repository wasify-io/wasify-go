package main

func main() {}

//go:wasmimport empty_host_func hostFunc
func hostFunc()

//export greet
func _greet() {
	hostFunc()
}
