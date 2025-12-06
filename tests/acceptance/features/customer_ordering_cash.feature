Feature: Customer Orders Food with Cash Payment
  In order to get food without digital payment friction
  As a hungry customer
  I need to be able to order food and pay with cash

  Scenario: Customer places a cash order
    Given "Sarah" visits the restaurant menu
    When she adds "Burger" to her cart
    And she places the order with "Cash"
    Then she should see the order status "UNPAID"

