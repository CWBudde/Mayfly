# GSASMA - Golden Annealing Crossover-Mutation Mayfly Algorithm

## Research reference

M. Zhao, X. Yang, and X. Yin, “An improved mayfly algorithm and its
application,” *AIP Advances* 12, 105320 (2022).
[DOI 10.1063/5.0108278](https://doi.org/10.1063/5.0108278)

## Implemented lifecycle

GSASMA changes the velocity and position stages for both male and female
populations.

During the first half of the run (`2*iteration < MaxIterations`), it uses the
ordinary Mayfly attraction/dance and attraction/random-flight velocity
branches. During the second half, it compares each mayfly's current fitness
with that same mayfly's previous fitness. An improvement selects attraction.
Otherwise, attraction is selected with the Metropolis probability

```text
eta = exp(-(f_current - f_previous) / T)
```

and the dance or random-flight velocity is selected when that draw fails.
Previous fitness is attached to mayfly identity, so sorting and survivor
selection cannot assign it to a different individual.

After selecting and clamping velocity, the library moves the mayfly and applies
the paper's Eq. (10) golden-sine refinement:

```text
z = clamp(x + v)
x_new = z*abs(sin(r1)) - r2*sin(r1)*abs(c1*pbest - c2*z)

r1 in [0, 2*pi]
r2 in [0, pi]
tau = (sqrt(5)-1)/2
c1 = -pi + (1-tau)*2*pi
c2 = -pi + tau*2*pi
```

The paper presents velocity and position improvements as consecutive components
but prints Eq. (10) using `x(t)`. The explicit `x+v` then golden-sine
composition is the necessary interpretation that keeps its annealed velocity
stage behaviorally effective.

## Configuration

```go
config := mayfly.NewGSASMAConfig()
config.ObjectiveFunc = mayfly.Rastrigin
config.ProblemSize = 30
config.LowerBound = -5.12
config.UpperBound = 5.12

result, err := mayfly.Optimize(config)
```

`InitialTemperature`, `CoolingRate`, and `CoolingSchedule` control the
library's annealing-temperature recurrence. The paper defines Metropolis
acceptance using temperature `T` and prints
`tau = max(tau_i, rand), i = 0,1,...,10`, calling `tau` a cooling coefficient.
It does not define `tau_i`, connect `tau` to `T`, give an initial temperature,
or state a temperature-update recurrence. The exponential, linear, and
logarithmic schedules are therefore implementation extensions, not paper
parameters. The configured initial temperature is held constant during the
ordinary-MA first half; the schedule starts advancing only when the annealed
second-half velocity phase begins.

`GoldenFactor` is deprecated and ignored. Eq. (10) contains no such multiplier;
the golden coefficients are fixed and do not narrow across candidates.

## Remaining source ambiguity

The paper compares AMA, SMA, and ASMA mating policies and says SMA was selected
for GSASMA. It publishes the SMA equations but not values for `pc_min`,
`pc_max`, `pm_min`, or `pm_max`, including in its parameter table. The library
does not invent these values: GSASMA currently uses ordinary configured Mayfly
crossover and Gaussian mutation (`NC`, `NM`, `CrossoverGamma`, and `Mu`). Thus,
the velocity and position stages are equation-tested, while the exact SMA
mating policy cannot be reproduced from the cited article alone.

An August 2026 search of the DOI/title across GitHub, Zenodo, Mendeley Data,
Figshare, and OSF found no author implementation or public raw benchmark data.
The article's data-availability statement instead says supporting data are
available from the corresponding author upon reasonable request. Calibration
therefore requires that author material; aggregate means and standard
deviations from 30 unseeded runs cannot uniquely identify the missing values.

The completed audit is preserved in a
[machine-readable clarification request](../reference-data/gsasma-2022-clarification-request.json).
It pins five exact-algorithm blockers (the initial temperature, temperature
recurrence, `tau_i` sequence/lifecycle, four SMA probability bounds, and
fitness orientation) and two historical-comparison blockers (the seed/RNG
lifecycle and raw 30-run outputs). Its correspondence draft is send-ready but
remains explicitly `not_sent`; neither the artifact nor aggregate resemblance
is evidence that those gates have been resolved.

Dispatch readiness was verified on 2026-08-29: the restricted mail bridge was
connected, and a non-marking Inbox/Sent/Drafts search for the DOI, acronym,
exact title, and proposed recipient found no prior thread. The bridge can append
to Drafts but cannot send mail. Creating the prepared draft remains gated on
explicit outbound-correspondence authorization.

Hybrid Cauchy/Gaussian offspring mutation and periodic opposition are not
GSASMA stages. See [HMMA](hmma.md) for its distinct mutation cascade.
