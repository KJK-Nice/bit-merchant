Feature: Kitchen Staff Fulfills Orders
  In order to manage the kitchen workflow
  As a kitchen staff member
  I need to see orders and update their status

  Scenario: Kitchen staff processes a cash order
    Given "Marcus" is on the kitchen dashboard
    When a new order is placed with "Cash"
    And he marks the order as "Paid"
    Then the order status should be "Paid"
    When he marks the order as "Preparing"
    Then the order status should be "Preparing"
    When he marks the order as "Ready"
    Then the order status should be "Ready"

