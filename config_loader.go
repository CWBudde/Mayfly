package mayfly

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// ConfigPreset represents predefined configurations for common problem types.
type ConfigPreset string

const (
	PresetUnimodal          ConfigPreset = "unimodal"
	PresetMultimodal        ConfigPreset = "multimodal"
	PresetHighlyMultimodal  ConfigPreset = "highly_multimodal"
	PresetDeceptive         ConfigPreset = "deceptive"
	PresetNarrowValley      ConfigPreset = "narrow_valley"
	PresetHighDimensional   ConfigPreset = "high_dimensional"
	PresetFastConvergence   ConfigPreset = "fast_convergence"
	PresetStableConvergence ConfigPreset = "stable_convergence"
	PresetMultiObjective    ConfigPreset = "multi_objective"
)

// LoadConfigFromFile loads a Config from a JSON file.
// Note: ObjectiveFunc and Rand must be set separately as they cannot be serialized.
func LoadConfigFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := &Config{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate the loaded config
	if err := ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return config, nil
}

// SaveConfigToFile saves a Config to a JSON file.
// Note: ObjectiveFunc and Rand are not saved as they cannot be serialized.
func SaveConfigToFile(config *Config, path string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ValidateConfig checks if a configuration is valid and provides helpful error messages.
func ValidateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}

	// Check required fields (note: ObjectiveFunc can be nil if loaded from file)
	if config.ProblemSize <= 0 {
		return fmt.Errorf("problem_size must be positive (got %d)", config.ProblemSize)
	}

	if config.LowerBound >= config.UpperBound {
		return fmt.Errorf("lower_bound (%f) must be less than upper_bound (%f)",
			config.LowerBound, config.UpperBound)
	}

	if config.MaxIterations <= 0 {
		return fmt.Errorf("max_iterations must be positive (got %d)", config.MaxIterations)
	}

	if config.NPop <= 0 {
		return fmt.Errorf("npop must be positive (got %d)", config.NPop)
	}

	if config.NPopF <= 0 {
		return fmt.Errorf("npopf must be positive (got %d)", config.NPopF)
	}

	// Validate coefficient ranges
	if config.G < 0 || config.G > 1 {
		return fmt.Errorf("g (inertia weight) should be in [0,1] (got %f)", config.G)
	}

	if config.GDamp <= 0 {
		return fmt.Errorf("g_damp must be positive (got %f)", config.GDamp)
	}

	if config.A1 < 0 || config.A2 < 0 || config.A3 < 0 {
		return fmt.Errorf("learning coefficients (a1, a2, a3) must be non-negative")
	}

	if config.Beta <= 0 {
		return fmt.Errorf("beta must be positive (got %f)", config.Beta)
	}

	if config.Mu < 0 || config.Mu > 1 {
		return fmt.Errorf("mu (mutation rate) should be in [0,1] (got %f)", config.Mu)
	}

	// Validate variant-specific parameters
	if config.UseDESMA {
		if config.EliteCount < 0 {
			return fmt.Errorf("elite_count must be non-negative (got %d)", config.EliteCount)
		}
		if config.EnlargeFactor <= 0 || config.ReductionFactor <= 0 {
			return fmt.Errorf("enlarge_factor and reduction_factor must be positive")
		}
	}

	if config.UseOLCE {
		if config.OrthogonalFactor < 0 || config.OrthogonalFactor > 1 {
			return fmt.Errorf("orthogonal_factor should be in [0,1] (got %f)", config.OrthogonalFactor)
		}
		if config.ChaosFactor < 0 || config.ChaosFactor > 1 {
			return fmt.Errorf("chaos_factor should be in [0,1] (got %f)", config.ChaosFactor)
		}
	}

	if config.UseEOBBMA {
		if config.LevyAlpha <= 0 || config.LevyAlpha > 2 {
			return fmt.Errorf("levy_alpha should be in (0,2] (got %f)", config.LevyAlpha)
		}
		if config.LevyBeta <= 0 {
			return fmt.Errorf("levy_beta must be positive (got %f)", config.LevyBeta)
		}
		if config.OppositionRate < 0 || config.OppositionRate > 1 {
			return fmt.Errorf("opposition_rate should be in [0,1] (got %f)", config.OppositionRate)
		}
	}

	if config.UseMPMA {
		if config.MedianWeight < 0 || config.MedianWeight > 1 {
			return fmt.Errorf("median_weight should be in [0,1] (got %f)", config.MedianWeight)
		}
		validGravityTypes := map[string]bool{"linear": true, "exponential": true, "sigmoid": true}
		if !validGravityTypes[config.GravityType] {
			return fmt.Errorf("gravity_type must be 'linear', 'exponential', or 'sigmoid' (got '%s')", config.GravityType)
		}
	}

	if config.UseGSASMA {
		if config.InitialTemperature <= 0 {
			return fmt.Errorf("initial_temperature must be positive (got %f)", config.InitialTemperature)
		}
		if config.CoolingRate <= 0 || config.CoolingRate >= 1 {
			return fmt.Errorf("cooling_rate should be in (0,1) (got %f)", config.CoolingRate)
		}
		if config.CauchyMutationRate < 0 || config.CauchyMutationRate > 1 {
			return fmt.Errorf("cauchy_mutation_rate should be in [0,1] (got %f)", config.CauchyMutationRate)
		}
		validSchedules := map[string]bool{"exponential": true, "linear": true, "logarithmic": true}
		if !validSchedules[config.CoolingSchedule] {
			return fmt.Errorf("cooling_schedule must be 'exponential', 'linear', or 'logarithmic' (got '%s')", config.CoolingSchedule)
		}
	}

	if config.UseAOBLMOA {
		if config.AquilaWeight < 0 || config.AquilaWeight > 1 {
			return fmt.Errorf("aquila_weight should be in [0,1] (got %f)", config.AquilaWeight)
		}
		if config.OppositionProbability < 0 || config.OppositionProbability > 1 {
			return fmt.Errorf("opposition_probability should be in [0,1] (got %f)", config.OppositionProbability)
		}
		if config.ArchiveSize < 0 {
			return fmt.Errorf("archive_size must be non-negative (got %d)", config.ArchiveSize)
		}
	}

	// Check for conflicting variants
	activeVariants := 0
	if config.UseDESMA {
		activeVariants++
	}
	if config.UseOLCE {
		activeVariants++
	}
	if config.UseEOBBMA {
		activeVariants++
	}
	if config.UseMPMA {
		activeVariants++
	}
	if config.UseGSASMA {
		activeVariants++
	}
	if config.UseAOBLMOA {
		activeVariants++
	}

	if activeVariants > 1 {
		return fmt.Errorf("multiple algorithm variants enabled (only one can be active at a time)")
	}

	return nil
}

// NewPresetConfig creates a configuration based on a predefined preset for common problem types.
// You must still set ObjectiveFunc, ProblemSize, LowerBound, and UpperBound.
func NewPresetConfig(preset ConfigPreset) (*Config, error) {
	var config *Config

	switch preset {
	case PresetUnimodal:
		// Standard MA works well on unimodal problems
		config = NewDefaultConfig()

	case PresetMultimodal:
		// DESMA for multimodal problems
		config = NewDESMAConfig()

	case PresetHighlyMultimodal:
		// OLCE-MA for highly multimodal problems
		config = NewOLCEConfig()

	case PresetDeceptive:
		// EOBBMA for deceptive landscapes
		config = NewEOBBMAConfig()

	case PresetNarrowValley:
		// MPMA for ill-conditioned problems
		config = NewMPMAConfig()

	case PresetHighDimensional:
		// OLCE-MA with larger population
		config = NewOLCEConfig()
		config.NPop = 40
		config.NPopF = 40
		config.MaxIterations = 1000

	case PresetFastConvergence:
		// GSASMA for fast convergence
		config = NewGSASMAConfig()
		config.MaxIterations = 300 // Fewer iterations

	case PresetStableConvergence:
		// MPMA for stable convergence
		config = NewMPMAConfig()

	case PresetMultiObjective:
		// AOBLMOA for multi-objective
		config = NewAOBLMOAConfig()

	default:
		return nil, fmt.Errorf("unknown preset: %s", preset)
	}

	return config, nil
}

// ListPresets returns all available configuration presets with descriptions.
func ListPresets() map[ConfigPreset]string {
	return map[ConfigPreset]string{
		PresetUnimodal:          "Standard MA - For unimodal problems with single optimum",
		PresetMultimodal:        "DESMA - For multimodal problems with several local optima",
		PresetHighlyMultimodal:  "OLCE-MA - For highly multimodal problems with many local optima",
		PresetDeceptive:         "EOBBMA - For deceptive landscapes with misleading gradients",
		PresetNarrowValley:      "MPMA - For ill-conditioned problems with narrow valleys",
		PresetHighDimensional:   "OLCE-MA - For high-dimensional problems (20D+)",
		PresetFastConvergence:   "GSASMA - For problems requiring fast convergence",
		PresetStableConvergence: "MPMA - For problems requiring stable, robust convergence",
		PresetMultiObjective:    "AOBLMOA - For multi-objective optimization",
	}
}

// PrintPresets prints all available presets with descriptions.
func PrintPresets() {
	fmt.Println("Available Configuration Presets:")
	fmt.Println(strings.Repeat("=", 80))

	presets := ListPresets()
	for preset, description := range presets {
		fmt.Printf("  %-25s : %s\n", preset, description)
	}

	fmt.Println(strings.Repeat("=", 80))
}

// AutoTuneConfig performs basic auto-tuning of configuration parameters based on problem characteristics.
// This is a simple heuristic-based approach, not an exhaustive search.
func AutoTuneConfig(config *Config, characteristics ProblemCharacteristics) {
	// Adjust population size based on dimensionality
	if characteristics.Dimensionality >= 50 {
		config.NPop = 40
		config.NPopF = 40
	} else if characteristics.Dimensionality >= 20 {
		config.NPop = 30
		config.NPopF = 30
	}

	// Adjust iterations based on requirements
	if characteristics.RequiresFastConvergence {
		config.MaxIterations = max(config.MaxIterations/2, 200)
	} else if characteristics.Dimensionality >= 50 {
		config.MaxIterations = max(config.MaxIterations*2, 1000)
	}

	// Adjust variant-specific parameters
	if config.UseGSASMA {
		if characteristics.RequiresFastConvergence {
			config.CoolingRate = 0.90 // Faster cooling
		} else {
			config.CoolingRate = 0.98 // Slower cooling for thorough exploration
		}
	}

	if config.UseMPMA {
		if characteristics.Landscape == NarrowValley {
			config.GravityType = "sigmoid" // Smooth transition
		} else {
			config.GravityType = "exponential" // Faster exploitation
		}
	}

	if config.UseOLCE {
		if characteristics.Modality == HighlyMultimodal {
			config.OrthogonalFactor = 0.4 // Increase diversity
		}
	}

	if config.UseAOBLMOA {
		if characteristics.MultiObjective {
			config.ArchiveSize = 200 // Larger archive for multi-objective
		}
	}
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ExportConfigTemplate creates a template JSON configuration file with all parameters and comments.
func ExportConfigTemplate(path string, variant string) error {
	var config *Config

	// Create config based on variant
	v := NewVariant(variant)
	if v != nil {
		config = v.GetConfig()
	} else {
		config = NewDefaultConfig()
	}

	// Create a map to include comments (not supported by direct JSON marshal)
	// We'll create a formatted JSON manually with comments

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create template file: %w", err)
	}
	defer file.Close()

	// Write JSON with inline comments (JSON5 style, but parseable as standard JSON)
	fmt.Fprintf(file, "{\n")
	fmt.Fprintf(file, "  // Problem parameters\n")
	fmt.Fprintf(file, "  \"problem_size\": %d,\n", config.ProblemSize)
	fmt.Fprintf(file, "  \"lower_bound\": %f,\n", config.LowerBound)
	fmt.Fprintf(file, "  \"upper_bound\": %f,\n", config.UpperBound)
	fmt.Fprintf(file, "\n")
	fmt.Fprintf(file, "  // Algorithm parameters\n")
	fmt.Fprintf(file, "  \"max_iterations\": %d,\n", config.MaxIterations)
	fmt.Fprintf(file, "  \"npop\": %d,\n", config.NPop)
	fmt.Fprintf(file, "  \"npopf\": %d,\n", config.NPopF)
	fmt.Fprintf(file, "  \"g\": %f,\n", config.G)
	fmt.Fprintf(file, "  \"g_damp\": %f,\n", config.GDamp)
	fmt.Fprintf(file, "  \"a1\": %f,\n", config.A1)
	fmt.Fprintf(file, "  \"a2\": %f,\n", config.A2)
	fmt.Fprintf(file, "  \"a3\": %f,\n", config.A3)
	fmt.Fprintf(file, "  \"beta\": %f,\n", config.Beta)
	fmt.Fprintf(file, "  \"dance\": %f,\n", config.Dance)
	fmt.Fprintf(file, "  \"fl\": %f,\n", config.FL)
	fmt.Fprintf(file, "  \"dance_damp\": %f,\n", config.DanceDamp)
	fmt.Fprintf(file, "  \"fl_damp\": %f,\n", config.FLDamp)
	fmt.Fprintf(file, "\n")
	fmt.Fprintf(file, "  // Mating parameters\n")
	fmt.Fprintf(file, "  \"nc\": %d,\n", config.NC)
	fmt.Fprintf(file, "  \"nm\": %d,\n", config.NM)
	fmt.Fprintf(file, "  \"mu\": %f,\n", config.Mu)
	fmt.Fprintf(file, "\n")
	fmt.Fprintf(file, "  // Velocity limits (0 = auto-calculated)\n")
	fmt.Fprintf(file, "  \"vel_max\": %f,\n", config.VelMax)
	fmt.Fprintf(file, "  \"vel_min\": %f,\n", config.VelMin)
	fmt.Fprintf(file, "\n")
	fmt.Fprintf(file, "  // Variant flags (only one should be true)\n")
	fmt.Fprintf(file, "  \"use_desma\": %t,\n", config.UseDESMA)
	fmt.Fprintf(file, "  \"use_olce\": %t,\n", config.UseOLCE)
	fmt.Fprintf(file, "  \"use_eobbma\": %t,\n", config.UseEOBBMA)
	fmt.Fprintf(file, "  \"use_mpma\": %t,\n", config.UseMPMA)
	fmt.Fprintf(file, "  \"use_gsasma\": %t,\n", config.UseGSASMA)
	fmt.Fprintf(file, "  \"use_aoblmoa\": %t\n", config.UseAOBLMOA)
	fmt.Fprintf(file, "}\n")

	fmt.Fprintf(file, "\n// Note: This template contains comments for readability.\n")
	fmt.Fprintf(file, "// Remove comments before loading with LoadConfigFromFile().\n")

	return nil
}
