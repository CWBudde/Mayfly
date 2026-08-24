package mayfly

import "math"

// EngineeringBenchmarkSuite returns four standard constrained engineering
// design problems. Objectives and constraints use physical coordinates; call
// BenchmarkCase.NewConfig to optimize them in Mayfly's unit search box.
func EngineeringBenchmarkSuite() ([]*BenchmarkCase, error) {
	constructors := []func() (*BenchmarkCase, error){
		NewSpringDesignProblem,
		NewWeldedBeamDesignProblem,
		NewPressureVesselDesignProblem,
		NewSpeedReducerDesignProblem,
	}

	problems := make([]*BenchmarkCase, 0, len(constructors))
	for _, constructor := range constructors {
		problem, err := constructor()
		if err != nil {
			return nil, err
		}

		problems = append(problems, problem)
	}

	return problems, nil
}

// NewSpringDesignProblem returns the constrained tension/compression spring
// weight problem. Coordinates are wire diameter, mean coil diameter, and the
// number of active coils.
func NewSpringDesignProblem() (*BenchmarkCase, error) {
	objective := func(x []float64) float64 {
		return (x[2] + 2) * x[1] * x[0] * x[0]
	}
	constraints := &ConstraintConfig{
		Handling: ConstraintHandlingFeasibility,
		Inequalities: []ConstraintFunction{
			func(x []float64) float64 { return 1 - x[1]*x[1]*x[1]*x[2]/(71785*x[0]*x[0]*x[0]*x[0]) },
			func(x []float64) float64 {
				numerator := 4*x[1]*x[1] - x[0]*x[1]
				denominator := 12566 * (x[1]*x[0]*x[0]*x[0] - x[0]*x[0]*x[0]*x[0])

				return numerator/denominator + 1/(5108*x[0]*x[0]) - 1
			},
			func(x []float64) float64 { return 1 - 140.45*x[0]/(x[1]*x[1]*x[2]) },
			func(x []float64) float64 { return (x[0]+x[1])/1.5 - 1 },
		},
	}

	return newBenchmarkCase(BenchmarkCase{
		suite: "Engineering", name: "Tension/compression spring design", dimension: 3,
		lower: []float64{0.05, 0.25, 2}, upper: []float64{2, 1.3, 15},
		optimum: []float64{0.051689031917057, 0.356717038149551, 11.289006887322081}, minimum: 0.012665232788377,
		objective: objective, constraints: constraints,
	})
}

// NewWeldedBeamDesignProblem returns the welded-beam fabrication-cost
// problem. Coordinates are weld thickness, weld length, beam height, and beam
// thickness, in inches.
func NewWeldedBeamDesignProblem() (*BenchmarkCase, error) {
	objective := func(x []float64) float64 {
		return 1.10471*x[0]*x[0]*x[1] + 0.04811*x[2]*x[3]*(14+x[1])
	}
	constraints := &ConstraintConfig{
		Handling: ConstraintHandlingFeasibility,
		Inequalities: []ConstraintFunction{
			func(x []float64) float64 { return weldedBeamShear(x)/13600 - 1 },
			func(x []float64) float64 { return weldedBeamStress(x)/30000 - 1 },
			func(x []float64) float64 { return x[0]/x[3] - 1 },
			func(x []float64) float64 {
				return (0.10471*x[0]*x[0]+0.04811*x[2]*x[3]*(14+x[1]))/5 - 1
			},
			func(x []float64) float64 { return 0.125/x[0] - 1 },
			func(x []float64) float64 { return weldedBeamDeflection(x)/0.25 - 1 },
			func(x []float64) float64 { return 6000/weldedBeamBucklingLoad(x) - 1 },
		},
	}

	return newBenchmarkCase(BenchmarkCase{
		suite: "Engineering", name: "Welded beam design", dimension: 4,
		lower: []float64{0.1, 0.1, 0.1, 0.1}, upper: []float64{2, 10, 10, 2},
		optimum: []float64{
			0.205729639786079, 3.470488665628002,
			9.036623910357633, 0.205729639786080,
		},
		minimum:   1.724852308597365,
		objective: objective, constraints: constraints,
	})
}

func weldedBeamShear(x []float64) float64 {
	primary := 6000 / (math.Sqrt2 * x[0] * x[1])
	moment := 6000 * (14 + x[1]/2)
	radius := math.Sqrt(x[1]*x[1]/4 + (x[0]+x[2])/2*((x[0]+x[2])/2))
	polar := 2 * math.Sqrt2 * x[0] * x[1] * (x[1]*x[1]/12 + (x[0]+x[2])/2*((x[0]+x[2])/2))
	secondary := moment * radius / polar

	return math.Sqrt(primary*primary + primary*secondary*x[1]/radius + secondary*secondary)
}

func weldedBeamStress(x []float64) float64 {
	return 6 * 6000 * 14 / (x[3] * x[2] * x[2])
}

func weldedBeamDeflection(x []float64) float64 {
	return 4 * 6000 * float64(14*14*14) / (30e6 * x[3] * x[2] * x[2] * x[2])
}

func weldedBeamBucklingLoad(x []float64) float64 {
	return 4.013 * 30e6 * math.Sqrt(x[2]*x[2]*math.Pow(x[3], 6)/36) / (14 * 14) *
		(1 - x[2]/(2*14)*math.Sqrt(30e6/(4*12e6)))
}

// NewPressureVesselDesignProblem returns the mixed-variable pressure-vessel
// cost problem. The first two coordinates are integer multiples of 0.0625 inch
// for shell and head thickness; they are rounded to the nearest integer index
// before evaluation. The remaining coordinates are radius and cylinder length.
func NewPressureVesselDesignProblem() (*BenchmarkCase, error) {
	optimumRadius := 0.8125 / 0.0193
	optimumLength := (1296000 - 4*math.Pi*optimumRadius*optimumRadius*optimumRadius/3) /
		(math.Pi * optimumRadius * optimumRadius)
	project := func(x []float64) []float64 {
		result := append([]float64(nil), x...)
		result[0] = math.Round(result[0])
		result[1] = math.Round(result[1])

		return result
	}
	physicalThickness := func(x []float64) (float64, float64) { return 0.0625 * x[0], 0.0625 * x[1] }
	objective := func(x []float64) float64 {
		shell, head := physicalThickness(x)

		return 0.6224*shell*x[2]*x[3] + 1.7781*head*x[2]*x[2] +
			3.1661*shell*shell*x[3] + 19.84*shell*shell*x[2]
	}
	constraints := &ConstraintConfig{
		Handling: ConstraintHandlingFeasibility,
		Inequalities: []ConstraintFunction{
			func(x []float64) float64 { shell, _ := physicalThickness(x); return 0.0193*x[2]/shell - 1 },
			func(x []float64) float64 { _, head := physicalThickness(x); return 0.00954*x[2]/head - 1 },
			func(x []float64) float64 {
				volume := math.Pi*x[2]*x[2]*x[3] + 4*math.Pi*x[2]*x[2]*x[2]/3
				return 1 - volume/1296000
			},
			func(x []float64) float64 { return x[3]/240 - 1 },
		},
	}

	return newBenchmarkCase(BenchmarkCase{
		suite: "Engineering", name: "Pressure vessel design", dimension: 4,
		lower: []float64{1, 1, 10, 10}, upper: []float64{99, 99, 200, 200},
		optimum: []float64{13, 7, optimumRadius, optimumLength}, minimum: 6059.714335048453,
		objective: objective, constraints: constraints, project: project,
	})
}

// NewSpeedReducerDesignProblem returns the mixed-variable seven-parameter
// speed-reducer weight problem. Gear-tooth count (coordinate 3) is rounded to
// the nearest integer before evaluation.
func NewSpeedReducerDesignProblem() (*BenchmarkCase, error) {
	project := func(x []float64) []float64 {
		result := append([]float64(nil), x...)
		result[2] = math.Round(result[2])

		return result
	}
	objective := func(x []float64) float64 {
		return 0.7854*x[0]*x[1]*x[1]*(3.3333*x[2]*x[2]+14.9334*x[2]-43.0934) -
			1.508*x[0]*(x[5]*x[5]+x[6]*x[6]) + 7.4777*(x[5]*x[5]*x[5]+x[6]*x[6]*x[6]) +
			0.7854*(x[3]*x[5]*x[5]+x[4]*x[6]*x[6])
	}
	constraints := &ConstraintConfig{
		Handling: ConstraintHandlingFeasibility,
		Inequalities: []ConstraintFunction{
			func(x []float64) float64 { return 27/(x[0]*x[1]*x[1]*x[2]) - 1 },
			func(x []float64) float64 { return 397.5/(x[0]*x[1]*x[1]*x[2]*x[2]) - 1 },
			func(x []float64) float64 { return 1.93*x[3]*x[3]*x[3]/(x[1]*x[2]*math.Pow(x[5], 4)) - 1 },
			func(x []float64) float64 { return 1.93*x[4]*x[4]*x[4]/(x[1]*x[2]*math.Pow(x[6], 4)) - 1 },
			func(x []float64) float64 {
				return math.Sqrt(745*x[3]/(x[1]*x[2])*(745*x[3]/(x[1]*x[2]))+16.9e6)/(110*x[5]*x[5]*x[5]) - 1
			},
			func(x []float64) float64 {
				return math.Sqrt(745*x[4]/(x[1]*x[2])*(745*x[4]/(x[1]*x[2]))+157.5e6)/(85*x[6]*x[6]*x[6]) - 1
			},
			func(x []float64) float64 { return x[1]*x[2]/40 - 1 },
			func(x []float64) float64 { return 5*x[1]/x[0] - 1 },
			func(x []float64) float64 { return x[0]/(12*x[1]) - 1 },
			func(x []float64) float64 { return (1.5*x[5]+1.9)/x[3] - 1 },
			func(x []float64) float64 { return (1.1*x[6]+1.9)/x[4] - 1 },
		},
	}

	return newBenchmarkCase(BenchmarkCase{
		suite: "Engineering", name: "Speed reducer design", dimension: 7,
		lower: []float64{2.6, 0.7, 17, 7.3, 7.3, 2.9, 5},
		upper: []float64{3.6, 0.8, 28, 8.3, 8.3, 3.9, 5.5},
		optimum: []float64{
			3.5, 0.7, 17, 7.3, 7.715319912497795,
			3.350214666225438, 5.286654465026051,
		},
		minimum:   2994.471066247639,
		objective: objective, constraints: constraints, project: project,
	})
}
