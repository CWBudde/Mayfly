Feature: Constraint Handling
  As a user of the Mayfly library
  I want constraints to participate in candidate selection
  So that optimization returns feasible solutions without hiding raw objective costs

  Scenario: Feasibility rules prefer a feasible solution
    Given a constrained one-dimensional problem
    When I optimize it using feasibility rules
    Then the returned solution should satisfy all configured constraints
    And the reported cost should be the raw objective cost
