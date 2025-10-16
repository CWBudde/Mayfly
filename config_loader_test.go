package mayfly

import (
	"os"
	"testing"
)

// =============================================================================
// Tests for config_loader.go - Configuration Management
// =============================================================================

func TestSaveAndLoadConfig(t *testing.T) {
	// Create a config
	config := NewDESMAConfig()
	config.ProblemSize = 15
	config.LowerBound = -5
	config.UpperBound = 5
	config.MaxIterations = 300
	config.EliteCount = 7

	// Save to temp file
	tmpFile := "/tmp/test_mayfly_config.json"
	defer os.Remove(tmpFile)

	err := SaveConfigToFile(config, tmpFile)
	if err != nil {
		t.Fatalf("SaveConfigToFile failed: %v", err)
	}

	// Load it back
	loadedConfig, err := LoadConfigFromFile(tmpFile)
	if err != nil {
		t.Fatalf("LoadConfigFromFile failed: %v", err)
	}

	// Compare key fields
	if loadedConfig.ProblemSize != 15 {
		t.Errorf("Expected ProblemSize 15, got %d", loadedConfig.ProblemSize)
	}

	if loadedConfig.LowerBound != -5 {
		t.Errorf("Expected LowerBound -5, got %f", loadedConfig.LowerBound)
	}

	if loadedConfig.UpperBound != 5 {
		t.Errorf("Expected UpperBound 5, got %f", loadedConfig.UpperBound)
	}

	if loadedConfig.MaxIterations != 300 {
		t.Errorf("Expected MaxIterations 300, got %d", loadedConfig.MaxIterations)
	}

	if loadedConfig.EliteCount != 7 {
		t.Errorf("Expected EliteCount 7, got %d", loadedConfig.EliteCount)
	}

	if !loadedConfig.UseDESMA {
		t.Error("Expected UseDESMA to be true")
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		config  *Config
		name    string
		wantErr bool
	}{
		{
			name:    "Nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name: "Valid config",
			config: &Config{
				ProblemSize:   10,
				LowerBound:    -10,
				UpperBound:    10,
				MaxIterations: 100,
				NPop:          20,
				NPopF:         20,
				G:             0.8,
				GDamp:         1.0,
				A1:            1.0,
				A2:            1.5,
				A3:            1.5,
				Beta:          2.0,
				Mu:            0.01,
			},
			wantErr: false,
		},
		{
			name: "Invalid problem size",
			config: &Config{
				ProblemSize:   -1,
				LowerBound:    -10,
				UpperBound:    10,
				MaxIterations: 100,
				NPop:          20,
				NPopF:         20,
			},
			wantErr: true,
		},
		{
			name: "Invalid bounds",
			config: &Config{
				ProblemSize:   10,
				LowerBound:    10,
				UpperBound:    -10,
				MaxIterations: 100,
				NPop:          20,
				NPopF:         20,
			},
			wantErr: true,
		},
		{
			name: "Zero population",
			config: &Config{
				ProblemSize:   10,
				LowerBound:    -10,
				UpperBound:    10,
				MaxIterations: 100,
				NPop:          0,
				NPopF:         20,
			},
			wantErr: true,
		},
		{
			name: "Multiple variants enabled",
			config: &Config{
				ProblemSize:   10,
				LowerBound:    -10,
				UpperBound:    10,
				MaxIterations: 100,
				NPop:          20,
				NPopF:         20,
				G:             0.8,
				GDamp:         1.0,
				A1:            1.0,
				A2:            1.5,
				A3:            1.5,
				Beta:          2.0,
				Mu:            0.01,
				UseDESMA:      true,
				UseOLCE:       true, // Both enabled - invalid
			},
			wantErr: true,
		},
		{
			name: "Invalid MPMA gravity type",
			config: &Config{
				ProblemSize:   10,
				LowerBound:    -10,
				UpperBound:    10,
				MaxIterations: 100,
				NPop:          20,
				NPopF:         20,
				G:             0.8,
				GDamp:         1.0,
				A1:            1.0,
				A2:            1.5,
				A3:            1.5,
				Beta:          2.0,
				Mu:            0.01,
				UseMPMA:       true,
				GravityType:   "invalid", // Invalid type
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPresetConfigs(t *testing.T) {
	presets := []ConfigPreset{
		PresetUnimodal,
		PresetMultimodal,
		PresetHighlyMultimodal,
		PresetDeceptive,
		PresetNarrowValley,
		PresetHighDimensional,
		PresetFastConvergence,
		PresetStableConvergence,
		PresetMultiObjective,
	}

	for _, preset := range presets {
		t.Run(string(preset), func(t *testing.T) {
			config, err := NewPresetConfig(preset)
			if err != nil {
				t.Fatalf("NewPresetConfig(%s) failed: %v", preset, err)
			}

			if config == nil {
				t.Fatal("Config should not be nil")
			}

			// Set required fields for validation
			config.ObjectiveFunc = Sphere
			config.ProblemSize = 10
			config.LowerBound = -10
			config.UpperBound = 10

			err = ValidateConfig(config)
			if err != nil {
				t.Errorf("Preset config %s failed validation: %v", preset, err)
			}

			// Verify correct variant is enabled based on preset
			switch preset {
			case PresetUnimodal:
				if config.UseDESMA || config.UseOLCE || config.UseEOBBMA || config.UseMPMA || config.UseGSASMA || config.UseAOBLMOA {
					t.Error("Unimodal preset should use standard MA")
				}
			case PresetMultimodal:
				if !config.UseDESMA {
					t.Error("Multimodal preset should use DESMA")
				}
			case PresetHighlyMultimodal:
				if !config.UseOLCE {
					t.Error("Highly multimodal preset should use OLCE-MA")
				}
			case PresetDeceptive:
				if !config.UseEOBBMA {
					t.Error("Deceptive preset should use EOBBMA")
				}
			case PresetNarrowValley:
				if !config.UseMPMA {
					t.Error("Narrow valley preset should use MPMA")
				}
			case PresetFastConvergence:
				if !config.UseGSASMA {
					t.Error("Fast convergence preset should use GSASMA")
				}
			case PresetStableConvergence:
				if !config.UseMPMA {
					t.Error("Stable convergence preset should use MPMA")
				}
			case PresetMultiObjective:
				if !config.UseAOBLMOA {
					t.Error("Multi-objective preset should use AOBLMOA")
				}
			}
		})
	}
}

func TestListPresets(t *testing.T) {
	presets := ListPresets()

	// Should have 9 presets
	if len(presets) != 9 {
		t.Errorf("Expected 9 presets, got %d", len(presets))
	}

	// Each preset should have a description
	for preset, description := range presets {
		if description == "" {
			t.Errorf("Preset %s has empty description", preset)
		}
	}

	// Check for required presets
	requiredPresets := []ConfigPreset{
		PresetUnimodal,
		PresetMultimodal,
		PresetHighlyMultimodal,
		PresetDeceptive,
		PresetNarrowValley,
		PresetHighDimensional,
		PresetFastConvergence,
		PresetStableConvergence,
		PresetMultiObjective,
	}

	for _, preset := range requiredPresets {
		if _, found := presets[preset]; !found {
			t.Errorf("Required preset %s not found", preset)
		}
	}
}

func TestAutoTuneConfig(t *testing.T) {
	// Test high dimensional tuning
	config := NewDefaultConfig()
	originalPop := config.NPop

	characteristics := ProblemCharacteristics{
		Dimensionality: 60,
	}

	AutoTuneConfig(config, characteristics)

	// Population should be increased
	if config.NPop <= originalPop {
		t.Error("Population should be increased for high-dimensional problems")
	}

	// Test fast convergence tuning
	config = NewDefaultConfig()
	originalIterations := config.MaxIterations

	characteristics = ProblemCharacteristics{
		RequiresFastConvergence: true,
	}

	AutoTuneConfig(config, characteristics)

	if config.MaxIterations >= originalIterations {
		t.Error("Iterations should be reduced for fast convergence requirement")
	}

	// Test GSASMA-specific tuning
	config = NewGSASMAConfig()
	characteristics = ProblemCharacteristics{
		RequiresFastConvergence: true,
	}

	AutoTuneConfig(config, characteristics)

	// Should have faster cooling for fast convergence
	if config.CoolingRate >= 0.95 {
		t.Error("Cooling rate should be reduced for fast convergence")
	}

	// Test MPMA-specific tuning
	config = NewMPMAConfig()
	characteristics = ProblemCharacteristics{
		Landscape: NarrowValley,
	}

	AutoTuneConfig(config, characteristics)

	// Should use sigmoid gravity for narrow valley
	if config.GravityType != "sigmoid" {
		t.Errorf("Expected sigmoid gravity for narrow valley, got %s", config.GravityType)
	}
}

func TestInvalidPreset(t *testing.T) {
	_, err := NewPresetConfig("invalid_preset")
	if err == nil {
		t.Error("Should return error for invalid preset")
	}
}

func TestLoadInvalidConfigFile(t *testing.T) {
	// Test loading non-existent file
	_, err := LoadConfigFromFile("/tmp/nonexistent_file.json")
	if err == nil {
		t.Error("Should return error for non-existent file")
	}

	// Test loading invalid JSON
	tmpFile := "/tmp/test_invalid_config.json"
	defer os.Remove(tmpFile)

	err = os.WriteFile(tmpFile, []byte("invalid json {{{"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err = LoadConfigFromFile(tmpFile)
	if err == nil {
		t.Error("Should return error for invalid JSON")
	}
}
