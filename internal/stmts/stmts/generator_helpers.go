package stmts

import (
	"sqlfuse/internal/common"
)

// GeneratorConfig holds common configuration for statement generators.
// This applies the Null Object pattern to eliminate repetitive nil checks
// and provide sensible defaults.
type GeneratorConfig struct {
	LCG    *common.LCG
	Flavor FlavorConfig
}

// NewGeneratorConfig creates a new generator configuration with defaults.
// Any nil values are replaced with appropriate defaults:
// - LCG defaults to seed 1
// - Flavor defaults to DefaultFlavorConfig
func NewGeneratorConfig(lcg *common.LCG, flavor FlavorConfig) *GeneratorConfig {
	if lcg == nil {
		lcg = common.NewLCG(1)
	}
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}
	return &GeneratorConfig{
		LCG:    lcg,
		Flavor: flavor,
	}
}

// ensureLCG returns the provided LCG or creates a default one if nil.
// This is a convenience function for backward compatibility with existing code.
func ensureLCG(lcg *common.LCG) *common.LCG {
	if lcg == nil {
		return common.NewLCG(1)
	}
	return lcg
}

// ensureFlavor returns the provided flavor or creates a default one if nil.
// This is a convenience function for backward compatibility with existing code.
func ensureFlavor(flavor FlavorConfig) FlavorConfig {
	if flavor == nil {
		return GetDefaultFlavor()
	}
	return flavor
}
