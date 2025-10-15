Feature: Boundary Constraints
  As a user of the Mayfly library
  I want all solutions to respect boundary constraints
  So that optimization stays within valid search space

  Scenario: Positions remain within bounds during optimization
    Given a Rosenbrock function with dimension 5
    And bounds from -5 to 10
    When I run Standard MA for 200 iterations
    Then all male positions should be within bounds
    And all female positions should be within bounds

  Scenario: Offspring respect boundaries after genetic operations
    Given a Sphere function with dimension 10
    And bounds from -100 to 100
    When I run DESMA for 100 iterations
    Then all offspring positions should be within bounds after crossover
    And all offspring positions should be within bounds after mutation

  Scenario: Velocities are clamped correctly
    Given a Rastrigin function with dimension 5
    And bounds from -5.12 to 5.12
    When I run Standard MA for 50 iterations
    Then all velocities should be within calculated velocity bounds

  Scenario: Elite solutions respect boundaries
    Given an Ackley function with dimension 5
    And bounds from -10 to 10
    When I run DESMA for 100 iterations
    Then all elite solutions should be within bounds
