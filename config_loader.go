package mayfly

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"sort"
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
	// PresetMultiObjective is retained only so callers receive a clear error
	// instead of a compile failure. Mayfly has no multi-objective optimizer.
	// Deprecated: no replacement is available yet.
	PresetMultiObjective ConfigPreset = "multi_objective"
)

// LoadConfigFromFile loads a Config from a JSON file.
// Note: ObjectiveFunc, constraint functions, and Rand must be set separately as
// they cannot be serialized.
func LoadConfigFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	err = rejectDuplicateJSONFields(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	config := &Config{}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	if err = decoder.Decode(config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err = ensureJSONEOF(decoder); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate the loaded config
	err = ValidateConfig(config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return config, nil
}

// SaveConfigToFile saves a Config to a JSON file.
// Note: ObjectiveFunc, constraint functions, and Rand are not saved because
// they cannot be serialized.
func SaveConfigToFile(config *Config, path string) error {
	if config == nil {
		return errors.New("config is nil")
	}
	if err := ValidateConfig(config); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	return writeJSONAtomic(config, path)
}

// ValidateConfig checks if a configuration is valid and provides helpful error messages.
func ValidateConfig(config *Config) error {
	if config == nil {
		return errors.New("config is nil")
	}

	value := reflect.ValueOf(*config)

	typeInfo := value.Type()
	for i := 0; i < value.NumField(); i++ {
		if value.Field(i).Kind() != reflect.Float64 {
			continue
		}

		number := value.Field(i).Float()
		if math.IsNaN(number) || math.IsInf(number, 0) {
			name := strings.Split(typeInfo.Field(i).Tag.Get("json"), ",")[0]
			return fmt.Errorf("%s must be finite (got %v)", name, number)
		}
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

	convergenceErr := validateConvergenceConfig(config.Convergence, config.MaxIterations)
	if convergenceErr != nil {
		return fmt.Errorf("invalid convergence config: %w", convergenceErr)
	}

	constraintErr := validateConstraintConfig(config.Constraints)
	if constraintErr != nil {
		return fmt.Errorf("invalid constraint config: %w", constraintErr)
	}

	if config.NPop <= 0 {
		return fmt.Errorf("npop must be positive (got %d)", config.NPop)
	}

	if config.NPopF <= 0 {
		return fmt.Errorf("npopf must be positive (got %d)", config.NPopF)
	}

	if config.MaxWorkers < 0 {
		return fmt.Errorf("max_workers must be non-negative (got %d)", config.MaxWorkers)
	}

	if config.Seed != nil && config.Rand != nil {
		return errors.New("seed and Rand are mutually exclusive")
	}

	if (config.VelMax == 0) != (config.VelMin == 0) {
		return errors.New("vel_min and vel_max must both be zero (automatic) or both be explicit")
	}

	if config.VelMax != 0 && config.VelMin >= config.VelMax {
		return fmt.Errorf("vel_min (%f) must be less than vel_max (%f)", config.VelMin, config.VelMax)
	}

	// The pairing, mating and mutation-rate checks live beside Optimize's
	// because it needs them too, but a configuration loaded from a file has to
	// fail here rather than at the start of a run that its caller believed was
	// already validated. Mu used to be checked separately just below, which
	// left the file path guarded and the programmatic path panicking. They run
	// in the same order as in Optimize so both entry points report the same
	// error for the same configuration.
	pairingErr := validateFemalePairing(config)
	if pairingErr != nil {
		return pairingErr
	}

	offspringErr := validateOffspring(config)
	if offspringErr != nil {
		return offspringErr
	}

	// Validate coefficient ranges
	if config.G < 0 || config.G > 1 {
		return fmt.Errorf("g (inertia weight) should be in [0,1] (got %f)", config.G)
	}

	if config.GDamp <= 0 {
		return fmt.Errorf("g_damp must be positive (got %f)", config.GDamp)
	}

	if config.A1 < 0 || config.A2 < 0 || config.A3 < 0 {
		return errors.New("learning coefficients (a1, a2, a3) must be non-negative")
	}

	if config.Beta <= 0 {
		return fmt.Errorf("beta must be positive (got %f)", config.Beta)
	}

	err := validateQMCInit(config)
	if err != nil {
		return err
	}

	// Validate variant-specific parameters
	if config.UseDESMA {
		if config.EliteCount < 0 {
			return fmt.Errorf("elite_count must be non-negative (got %d)", config.EliteCount)
		}

		if config.EnlargeFactor <= 0 || config.ReductionFactor <= 0 {
			return errors.New("enlarge_factor and reduction_factor must be positive")
		}
	}

	if config.UseOLCE {
		if !isFinite(config.OrthogonalFactor) || config.OrthogonalFactor < 0 || config.OrthogonalFactor > 1 {
			return fmt.Errorf("orthogonal_factor should be in [0,1] (got %f)", config.OrthogonalFactor)
		}

		if !isFinite(config.ChaosFactor) || config.ChaosFactor < 0 || config.ChaosFactor > 1 {
			return fmt.Errorf("chaos_factor should be in [0,1] (got %f)", config.ChaosFactor)
		}
	}

	if config.UseEOBBMA {
		if config.LevyAlpha <= 0 || config.LevyAlpha >= 2 {
			return fmt.Errorf("levy_alpha should be in (0,2) for Mantegna sampling (got %f)", config.LevyAlpha)
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

		validGravityTypes := map[string]bool{
			GravityPaper: true, GravityLinear: true, GravityExponential: true, GravitySigmoid: true,
		}
		if !validGravityTypes[config.GravityType] {
			return fmt.Errorf("gravity_type must be 'paper', 'linear', 'exponential', or 'sigmoid' (got '%s')", config.GravityType)
		}
	}

	if config.UseGSASMA {
		if config.InitialTemperature <= 0 {
			return fmt.Errorf("initial_temperature must be positive (got %f)", config.InitialTemperature)
		}

		if config.CoolingRate <= 0 || config.CoolingRate >= 1 {
			return fmt.Errorf("cooling_rate should be in (0,1) (got %f)", config.CoolingRate)
		}

		validSchedules := map[string]bool{CoolingExponential: true, CoolingLinear: true, CoolingLogarithmic: true}
		if !validSchedules[config.CoolingSchedule] {
			return fmt.Errorf("cooling_schedule must be 'exponential', 'linear', or 'logarithmic' (got '%s')",
				config.CoolingSchedule)
		}
	}

	if config.UseHMMA {
		if !isFinite(config.HMMAInformationExchange) || config.HMMAInformationExchange <= 0 {
			return fmt.Errorf("hmma_information_exchange must be positive (got %f)",
				config.HMMAInformationExchange)
		}
		if !isFinite(config.HMMAScheduleOffset) ||
			config.HMMAScheduleOffset < 0 || config.HMMAScheduleOffset > 1 {
			return fmt.Errorf("hmma_schedule_offset should be in [0,1] (got %f)",
				config.HMMAScheduleOffset)
		}
		if !isFinite(config.HMMAArtificialMutation) ||
			config.HMMAArtificialMutation < 0 || config.HMMAArtificialMutation > 1 {
			return fmt.Errorf("hmma_artificial_mutation should be in [0,1] (got %f)",
				config.HMMAArtificialMutation)
		}
	}

	if config.UseAOBLMOA {
		if config.AquilaWeight != AquilaWeightAuto &&
			(config.AquilaWeight < 0 || config.AquilaWeight > 1) {
			return fmt.Errorf("aquila_weight should be in [0,1] or AquilaWeightAuto (%v) (got %f)",
				AquilaWeightAuto, config.AquilaWeight)
		}

		if config.OppositionProbability < 0 || config.OppositionProbability > 1 {
			return fmt.Errorf("opposition_probability should be in [0,1] (got %f)", config.OppositionProbability)
		}

		if config.ArchiveSize < 0 {
			return fmt.Errorf("archive_size must be non-negative (got %d)", config.ArchiveSize)
		}

		// Legal at or beyond max_iterations: the run then never leaves the
		// Aquila exploration phase.
		if config.StrategySwitch < 0 {
			return fmt.Errorf("strategy_switch must be non-negative (got %d)", config.StrategySwitch)
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

	if config.UseHMMA {
		activeVariants++
	}

	if config.UseAOBLMOA {
		activeVariants++
	}

	if activeVariants > 1 {
		return errors.New("multiple algorithm variants enabled (only one can be active at a time)")
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
	}
}

// PrintPresets prints all available presets with descriptions.
// Deprecated: use WritePresets to receive writer errors.
func PrintPresets() {
	_ = WritePresets(os.Stdout)
}

// WritePresets writes the available presets in deterministic name order.
func WritePresets(w io.Writer) error {
	if w == nil {
		return errors.New("preset writer is nil")
	}
	presets := ListPresets()
	names := make([]string, 0, len(presets))
	for preset := range presets {
		names = append(names, string(preset))
	}
	sort.Strings(names)
	var builder strings.Builder
	fmt.Fprintln(&builder, "Available Configuration Presets:")
	fmt.Fprintln(&builder, strings.Repeat("=", 80))
	for _, name := range names {
		fmt.Fprintf(&builder, "  %-25s : %s\n", name, presets[ConfigPreset(name)])
	}
	fmt.Fprintln(&builder, strings.Repeat("=", 80))
	_, err := io.WriteString(w, builder.String())
	return err
}

// AutoTuneConfig performs basic auto-tuning of configuration parameters based on problem characteristics.
// This is a simple heuristic-based approach, not an exhaustive search.
//
// Deprecated: use AutoTuneConfigChecked. This compatibility wrapper ignores a
// nil config and invalid problem metadata.
func AutoTuneConfig(config *Config, characteristics ProblemCharacteristics) {
	_ = AutoTuneConfigChecked(config, characteristics)
}

// AutoTuneConfigChecked applies the tuning heuristics after validating the
// mutable target and characteristic enum values.
func AutoTuneConfigChecked(config *Config, characteristics ProblemCharacteristics) error {
	if config == nil {
		return errors.New("config is nil")
	}
	if characteristics.Dimensionality < 0 {
		return fmt.Errorf("problem dimensionality must be non-negative, got %d", characteristics.Dimensionality)
	}
	if characteristics.Modality < Unimodal || characteristics.Modality > HighlyMultimodal {
		return fmt.Errorf("unknown problem modality %d", characteristics.Modality)
	}
	if characteristics.Landscape < Smooth || characteristics.Landscape > NarrowValley {
		return fmt.Errorf("unknown problem landscape %d", characteristics.Landscape)
	}
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
			config.GravityType = GravitySigmoid // Smooth transition
		} else {
			config.GravityType = GravityExponential // Faster exploitation
		}
	}

	if config.UseOLCE {
		if characteristics.Modality == HighlyMultimodal {
			config.OrthogonalFactor = 0.4 // Increase diversity
		}
	}
	return nil
}

// ExportConfigTemplate writes a complete, strict-JSON configuration for a
// named variant. The result can be loaded without first removing comments or
// filling fields that the variant constructor left at their zero values.
func ExportConfigTemplate(path, variant string) error {
	v := NewVariant(variant)
	if v == nil {
		return fmt.Errorf("unknown variant: %s", variant)
	}

	config := v.GetConfig()
	if config == nil {
		return fmt.Errorf("variant %s returned a nil config", variant)
	}

	config.ProblemSize = 10
	config.LowerBound = -10
	config.UpperBound = 10

	return writeJSONAtomic(config, path)
}

func ensureJSONEOF(decoder *json.Decoder) error {
	var trailing any

	err := decoder.Decode(&trailing)
	if !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("multiple JSON values are not allowed")
		}

		return err
	}

	return nil
}

func rejectDuplicateJSONFields(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))

	var walk func() error

	walk = func() error {
		token, err := decoder.Token()
		if err != nil {
			return err
		}

		delim, ok := token.(json.Delim)
		if !ok {
			return nil
		}

		switch delim {
		case '{':
			seen := make(map[string]struct{})

			for decoder.More() {
				keyToken, keyErr := decoder.Token()
				if keyErr != nil {
					return keyErr
				}

				key, ok := keyToken.(string)
				if !ok {
					return errors.New("object key is not a string")
				}

				if _, exists := seen[key]; exists {
					return fmt.Errorf("duplicate field %q", key)
				}

				seen[key] = struct{}{}

				err := walk()
				if err != nil {
					return err
				}
			}

			_, err = decoder.Token()

			return err
		case '[':
			for decoder.More() {
				err := walk()
				if err != nil {
					return err
				}
			}

			_, err = decoder.Token()

			return err
		default:
			return fmt.Errorf("unexpected JSON delimiter %q", delim)
		}
	}

	return walk()
}

func writeJSONAtomic(value any, path string) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}

	data = append(data, '\n')

	dir := filepath.Dir(path)

	tmp, err := os.CreateTemp(dir, ".mayfly-config-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary JSON file: %w", err)
	}

	tmpName := tmp.Name()
	removeTemp := true

	defer func() {
		if removeTemp {
			_ = os.Remove(tmpName)
		}
	}()

	if err = tmp.Chmod(0o600); err == nil {
		_, err = tmp.Write(data)
	}

	if err == nil {
		err = tmp.Sync()
	}

	closeErr := tmp.Close()
	if err == nil {
		err = closeErr
	}

	if err != nil {
		return fmt.Errorf("write temporary JSON file: %w", err)
	}

	if err = os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("replace JSON file: %w", err)
	}

	removeTemp = false

	return nil
}
