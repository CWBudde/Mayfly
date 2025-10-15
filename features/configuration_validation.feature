Feature: Configuration Validation
  As a user of the Mayfly library
  I want configuration errors to be caught early
  So that I can fix issues before running optimization

  Scenario: Missing ObjectiveFunc causes error
    Given a new Config instance
    When I set ProblemSize to 5
    And I set LowerBound to -10
    And I set UpperBound to 10
    And I call Optimize without setting ObjectiveFunc
    Then it should return an error containing "ObjectiveFunc"

  Scenario: Missing ProblemSize causes error
    Given a new Config instance
    When I set ObjectiveFunc to Sphere
    And I set LowerBound to -10
    And I set UpperBound to 10
    And I call Optimize without setting ProblemSize
    Then it should return an error containing "ProblemSize"

  Scenario: Auto-calculated NM uses correct defaults
    Given a valid Config with NPop set to 100
    When I do not set NM manually
    And I run optimization
    Then NM should be auto-calculated to 5

  Scenario: Auto-calculated velocity bounds are correct
    Given a valid Config with bounds from -10 to 10
    When I do not set VelMax and VelMin manually
    And I run optimization
    Then VelMax should be approximately 2
    And VelMin should be approximately -2

  Scenario Outline: DESMA auto-calculates SearchRange
    Given a DESMA config with bounds from <lower> to <upper>
    When I do not set SearchRange manually
    And I run optimization
    Then SearchRange should be approximately <expected>

    Examples:
      | lower | upper | expected |
      | -10   | 10    | 2        |
      | -5    | 5     | 1        |
      | -100  | 100   | 20       |
