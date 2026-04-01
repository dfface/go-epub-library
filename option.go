// SPDX-License-Identifier: Apache-2.0

package epub

// ComplianceLevel controls strictness of EPUB validations.
type ComplianceLevel int

const (
	// LevelFlexible allows broader structures as long as package integrity is valid.
	LevelFlexible ComplianceLevel = iota
	// LevelEBPAJ enforces EBPAJ-oriented naming and directory conventions.
	LevelEBPAJ
	// LevelKADOKAWA enforces KADOKAWA-oriented naming and directory conventions.
	LevelKADOKAWA
)

// decodeConfig stores Decode options.
type decodeConfig struct {
	compliance ComplianceLevel
}

// DecodeOption mutates decode behavior.
type DecodeOption func(*decodeConfig)

// WithCompliance configures strictness level used by Decode.
func WithCompliance(level ComplianceLevel) DecodeOption {
	return func(cfg *decodeConfig) {
		cfg.compliance = level
	}
}

func defaultDecodeConfig() decodeConfig {
	return decodeConfig{compliance: LevelFlexible}
}
