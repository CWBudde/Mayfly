Feature: Algorithm Variants
  As a user of the Mayfly library
  I want to run different algorithm variants correctly
  So that I can choose the best algorithm for my problem

  Scenario: Standard MA runs without DESMA features
    Given a Standard MA config
    When I run optimization for 100 iterations
    Then elite generation should not be called
    And search range should not be tracked
    And UseDESMA flag should be false

  Scenario: DESMA generates elite solutions
    Given a DESMA config with EliteCount set to 5
    When I run optimization for 100 iterations
    Then elite solutions should be generated after selection
    And search range should be initialized
    And UseDESMA flag should be true

  Scenario: DESMA adapts search range on improvement
    Given a DESMA config for Sphere function
    And initial SearchRange of 2.0
    When I run optimization for 200 iterations
    Then search range should increase if improving
    And search range should decrease if stagnating

  Scenario: Both variants use same core operators
    Given a Standard MA config
    And a DESMA config
    When I run both optimizations
    Then both should use Crossover operator
    And both should use Mutate operator
    And both should sort populations by fitness

  Scenario: DESMA EliteCount controls elite generation
    Given a DESMA config with EliteCount set to 3
    When I run optimization for 50 iterations
    Then exactly 3 elite solutions should be generated per iteration
