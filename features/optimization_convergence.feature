Feature: Optimization Convergence
  As a user of the Mayfly library
  I want to ensure algorithms converge to known optimal solutions
  So that I can trust the optimization results

  Scenario: Sphere function converges to global minimum
    Given a Sphere function with dimension 5
    And bounds from -10 to 10
    When I run Standard MA for 500 iterations
    Then the best cost should be less than 0.00001
    And the best position should be near zero vector within tolerance 0.01

  Scenario: DESMA and Standard MA both converge on Rastrigin
    Given a Rastrigin function with dimension 10
    And bounds from -5.12 to 5.12
    When I run Standard MA for 500 iterations with seed 42
    And I run DESMA for 500 iterations with seed 42
    Then the best cost should be less than 50

  Scenario: Rosenbrock converges to optimum
    Given a Rosenbrock function with dimension 5
    And bounds from -5 to 10
    When I run DESMA for 1000 iterations
    Then the best cost should be less than 10
    And the best position should be near ones vector within tolerance 0.5

  Scenario: Ackley function reaches near-zero
    Given an Ackley function with dimension 8
    And bounds from -32.768 to 32.768
    When I run DESMA for 500 iterations
    Then the best cost should be less than 0.1
