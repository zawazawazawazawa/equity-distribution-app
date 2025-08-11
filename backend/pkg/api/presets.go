package api

// OpponentPreset represents available preset names for opponent ranges
type OpponentPreset string

const (
	// PLO4 presets
	PresetPLO4_SRP_BBCallVsUTG  OpponentPreset = "SRP BB call vs UTG open"
	PresetPLO4_SRP_BBCallVsBTN  OpponentPreset = "SRP BB call vs BTN open"
	PresetPLO4_SRP_BTNCallVsUTG OpponentPreset = "SRP BTN call vs UTG open"
	PresetPLO4_3BP_UTGCallVsBB  OpponentPreset = "3BP UTG call vs BB 3bet"
	PresetPLO4_3BP_UTGCallVsBTN OpponentPreset = "3BP UTG call vs BTN 3bet"
	PresetPLO4_3BP_BTNCallVsBB  OpponentPreset = "3BP BTN call vs BB 3bet"

	// PLO5 presets
	PresetPLO5_SRP_BBCallVsUTG  OpponentPreset = "PLO5 SRP BB call vs UTG open"
	PresetPLO5_SRP_BBCallVsBTN  OpponentPreset = "PLO5 SRP BB call vs BTN open"
	PresetPLO5_SRP_BTNCallVsUTG OpponentPreset = "PLO5 SRP BTN call vs UTG open"
	PresetPLO5_3BP_UTGCallVsBB  OpponentPreset = "PLO5 3BP UTG call vs BB 3bet"
	PresetPLO5_3BP_UTGCallVsBTN OpponentPreset = "PLO5 3BP UTG call vs BTN 3bet"
	PresetPLO5_3BP_BTNCallVsBB  OpponentPreset = "PLO5 3BP BTN call vs BB 3bet"
)

// ValidPresets returns all valid preset values
func ValidPresets() []string {
	return []string{
		string(PresetPLO4_SRP_BBCallVsUTG),
		string(PresetPLO4_SRP_BBCallVsBTN),
		string(PresetPLO4_SRP_BTNCallVsUTG),
		string(PresetPLO4_3BP_UTGCallVsBB),
		string(PresetPLO4_3BP_UTGCallVsBTN),
		string(PresetPLO4_3BP_BTNCallVsBB),
		string(PresetPLO5_SRP_BBCallVsUTG),
		string(PresetPLO5_SRP_BBCallVsBTN),
		string(PresetPLO5_SRP_BTNCallVsUTG),
		string(PresetPLO5_3BP_UTGCallVsBB),
		string(PresetPLO5_3BP_UTGCallVsBTN),
		string(PresetPLO5_3BP_BTNCallVsBB),
	}
}

// IsValidPreset checks if a preset name is valid
func IsValidPreset(preset string) bool {
	for _, valid := range ValidPresets() {
		if preset == valid {
			return true
		}
	}
	return false
}