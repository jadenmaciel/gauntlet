# Copyable skeleton — godog v0.15.1 (cucumber/godog), standard Gherkin syntax.
# Replace the domain, scenario, and steps; keep the Given/When/Then shape.

Feature: Shipment carrier assignment
  As a fulfillment operator
  I want shipments with no carrier assigned to be rejected at dispatch
  So that no shipment leaves the warehouse without a carrier to hand it to

  Scenario: Reject a shipment with no carrier assigned
    Given a shipment "SHIP-1001" with no carrier assigned
    When the shipment is submitted for dispatch
    Then the dispatch is rejected
    And the rejection reason is "no carrier assigned"
