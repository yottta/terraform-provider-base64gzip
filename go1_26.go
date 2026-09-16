//go:build !go1.27

package main

// This ensures that that the provider will not be built with versions higher or equal with go1.27.
// The restriction is added because the whole purpose of this provider is to expose functions
// that use the underlying implementation from the go stdlib that changed started with go1.27.
// For more details, check the [README.md](./README.md).
const requiresGo1_26 = uint8(0)
