// This file provides abstractions and implementations for interacting with
// different WebAssembly runtimes, specifically focusing on the Wazero runtime.
package wasify

import (
	"context"
	"errors"
	"os"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// getWazeroRuntime creates and returns a wazero runtime instance using the provided context and
// RuntimeConfig. It configures the runtime with specific settings and features.
func getWazeroRuntime(ctx context.Context, c *RuntimeConfig) *wazeroRuntime {
	// TODO: Allow user to control the following options:
	// 1. WithCloseOnContextDone
	// 2. Memory
	// Create a new wazero runtime instance with specified configuration options.
	runtime := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().
		WithCoreFeatures(api.CoreFeaturesV2).
		WithCustomSections(false).
		WithCloseOnContextDone(false).
		// Enable runtime debug if user sets LogSeverity to debug level in runtime configuration
		WithDebugInfoEnabled(c.LogSeverity == LogDebug),
	)
	// Instantiate the runtime with the WASI snapshot preview1.
	wasi_snapshot_preview1.MustInstantiate(ctx, runtime)
	return &wazeroRuntime{runtime, c}
}

// The wazeroRuntime struct combines a wazero runtime instance with runtime configuration.
type wazeroRuntime struct {
	runtime wazero.Runtime
	*RuntimeConfig
}

// NewModule creates a new module instance based on the provided ModuleConfig within
//
// the wazero runtime context. It returns the created module and any potential error.
func (r *wazeroRuntime) NewModule(ctx context.Context, moduleConfig *ModuleConfig) (Module, error) {

	// Set the context, logger and any missing data for the moduleConfig.
	moduleConfig.ctx = ctx
	moduleConfig.log = r.log

	// Create a new wazeroModule instance and set its ModuleConfig.
	// Read more about wazeroModule in module_wazero.go
	wazeroModule := new(wazeroModule)
	wazeroModule.ModuleConfig = moduleConfig

	// If LogSeverity is set, create a new logger instance for the module.
	//
	// Module will adopt the log level from their parent runtime.
	// If you want only "Error" level for a runtime but need to debug specific module(s),
	// you can set those modules to "Debug". This will replace the inherited log level,
	// allowing the module to display debug information.
	if moduleConfig.LogSeverity != 0 {
		moduleConfig.log = newLogger(moduleConfig.LogSeverity)
	}

	// Check and compare hashes if provided in the moduleConfig.
	if moduleConfig.Wasm.Hash != "" {
		actualHash, err := calculateHash(moduleConfig.Wasm.Binary)
		if err != nil {
			err = errors.Join(errors.New("can't calculate the hash"), err)
			moduleConfig.log.Warn(err.Error(), "module", moduleConfig.Name, "needed hash", moduleConfig.Wasm.Hash, "actual wasm hash", actualHash)
			return nil, err
		}
		moduleConfig.log.Info("hash calculation", "module", moduleConfig.Name, "needed hash", moduleConfig.Wasm.Hash, "actual wasm hash", actualHash)

		err = compareHashes(actualHash, moduleConfig.Wasm.Hash)
		if err != nil {
			moduleConfig.log.Warn(err.Error(), "module", moduleConfig.Name, "needed hash", moduleConfig.Wasm.Hash, "actual wasm hash", actualHash)
			return nil, err
		}
	}

	// Instantiate host functions and configure wazeroModule accordingly.
	err := r.instantiateHostFunctions(ctx, wazeroModule, moduleConfig)
	if err != nil {
		moduleConfig.log.Error(err.Error(), "module", moduleConfig.Name)
		r.log.Error(err.Error(), "runtime", r.runtime, "module")
		return nil, err
	}

	moduleConfig.log.Info("host functions has been instantiated successfully", "module", moduleConfig.Name)

	// Instantiate the module and set it in wazeroModule.
	mod, err := r.instantiateModule(ctx, moduleConfig)
	if err != nil {
		moduleConfig.log.Error(err.Error(), "module", moduleConfig.Name)
		r.log.Error(err.Error(), "runtime", r.runtime, "module")
		return nil, err
	}

	moduleConfig.log.Info("module has been instantiated successfully", "module", moduleConfig.Name)

	wazeroModule.mod = mod

	return wazeroModule, nil
}

// convertToAPIValueTypes converts an array of ValueType values to their corresponding
// api.ValueType representations used by the Wazero runtime.
//
// ValueType describes a parameter or result type mapped to a WebAssembly
// function signature.
func (r *wazeroRuntime) convertToAPIValueTypes(types []ValueType) []api.ValueType {
	valueTypes := make([]api.ValueType, len(types))
	for i, t := range types {

		switch t {
		case ValueTypeByte:
			valueTypes[i] = api.ValueTypeI64
			// case ValueTypeI32:
			// 	valueTypes[i] = api.ValueTypeI32
			// case ValueTypeI64:
			// 	valueTypes[i] = api.ValueTypeI64
			// case ValueTypeF32:
			// 	valueTypes[i] = api.ValueTypeF32
			// case ValueTypeF64:
			// 	valueTypes[i] = api.ValueTypeF64
		}
	}

	return valueTypes
}

// instantiateHostFunctions sets up and exports host functions for the module using the wazero runtime.
//
// It configures host function callbacks, value types, and exports.
func (r *wazeroRuntime) instantiateHostFunctions(ctx context.Context, wazeroModule *wazeroModule, moduleConfig *ModuleConfig) error {

	modBuilder := r.runtime.NewHostModuleBuilder(moduleConfig.Name)

	// Iterate over the module's host functions and set up exports.
	for _, hf := range moduleConfig.HostFunctions {
		// Associate the host function with module-related information.
		// This configuration ensures that the host function can access ModuleConfig data from various contexts.
		// Additionally, we set up an allocationMap specific to the host function, creating a map that stores
		// offsets and sizes relevant to the host function's operations. This allows us to manage and clean up
		// user resources effectively.
		// We use allocationMap operations for Params provided in host function and Returns, which originally
		// should be freed up.
		// See host_function.go for more details.
		hf.moduleConfig = moduleConfig
		hf.allocationMap = newAllocationMap[uint32, uint32]()

		moduleConfig.log.Debug("exporting host function", "function", hf.Name, "module", moduleConfig.Name)

		modBuilder = modBuilder.
			NewFunctionBuilder().
			WithGoModuleFunction(api.GoModuleFunc(wazeroHostFunctionCallback(wazeroModule, moduleConfig, &hf)),
				r.convertToAPIValueTypes(hf.Params),
				r.convertToAPIValueTypes([]ValueType{ValueTypeByte}),
			).
			Export(hf.Name)
	}

	// modBuilder.
	// 	NewFunctionBuilder().
	// 	WithGoModuleFunction(api.GoModuleFunc(wazeroHostFunctionCallback(wazeroModule, moduleConfig, &log)),
	// 		r.convertToAPIValueTypes(log.Params),
	// 		r.convertToAPIValueTypes([]ValueType{ValueTypeByte}),
	// 	).
	// 	Export(log.Name)

	_, err := modBuilder.Instantiate(ctx)
	if err != nil {
		err = errors.Join(errors.New("can't instantiate NewHostModuleBuilder"), err)
		return err
	}

	return nil

}

// instantiateModule compiles and instantiates a WebAssembly module using the wazero runtime.
//
// It compiles the module, creates a module configuration, and then instantiates the module.
// Returns the instantiated module and any potential error.
func (r *wazeroRuntime) instantiateModule(ctx context.Context, moduleConfig *ModuleConfig) (api.Module, error) {

	// Compile the provided WebAssembly binary.
	compiled, err := r.runtime.CompileModule(ctx, moduleConfig.Wasm.Binary)
	if err != nil {
		return nil, errors.Join(errors.New("can't compile module"), err)
	}

	cfg := wazero.NewModuleConfig().
		// FIXME: The following line are added for simple debugging and should be removed to ensure sandboxing.
		WithStdout(os.Stdout).WithStderr(os.Stderr) // REMOVE LATER.

	if moduleConfig != nil && moduleConfig.FSConfig.Enabled {
		cfg = cfg.WithFSConfig(
			wazero.NewFSConfig().
				WithDirMount(moduleConfig.FSConfig.HostDir, moduleConfig.FSConfig.getGuestDir()),
		)
	}

	// Instantiate the compiled module with the provided module configuration.
	mod, err := r.runtime.InstantiateModule(ctx, compiled, cfg)
	if err != nil {
		return nil, errors.Join(errors.New("can't instantiate module"), err)
	}

	return mod, nil
}

// Close closes the resource.
//
// Note: The context parameter is used for value lookup, such as for
// logging. A canceled or otherwise done context will not prevent Close
// from succeeding.
func (r *wazeroRuntime) Close(ctx context.Context) error {
	err := r.runtime.Close(ctx)
	if err != nil {
		err = errors.Join(errors.New("can't close runtime"), err)
		r.log.Error(err.Error(), "module", r.runtime)
		return err
	}

	return nil
}
