# language: en
Feature: Complaint Filing BDD Tests
  As an AI agent using the complaints-mcp server
  I want to ensure that complaint filing works correctly across all scenarios
  
  Background:
    Given the MCP server is running
    And the file repository is properly initialized
    And the service layer is configured

  Scenario: File a valid complaint successfully
    Given I have valid complaint data
      | field         | value                     |
      |---------------|---------------------------|
      | task          | "Implement authentication system" |
      | context        | "Need to add JWT authentication to API endpoints" |
      | missing_info  | "Unclear error handling patterns" |
      | confused_by    | "Inconsistent response formats" |
      | future_wishes  | "Standardized error codes" |
      | severity       | "medium" |
      | agent_name     | "AI Assistant" |
      | project_name   | "user-auth-service" |
    And the MCP client is connected
    When I file the complaint using the file_complaint tool
    Then the complaint should be stored successfully
    And the complaint should have a valid ID
    And the complaint should be persisted to the file system
    And a success response should be returned

  Scenario: File a complaint with missing required field
    Given I have complaint data with missing task description
      | field         | value                     |
      |---------------|---------------------------|
      | task          | ""                         |
      | context        | "Some context about the task" |
      | missing_info  | "Need better documentation" |
      | confused_by    | "Response format unclear" |
      | future_wishes  | "More examples" |
      | severity       | "high" |
      | agent_name     | "AI Assistant" |
    And the MCP client is connected
    When I attempt to file the complaint
    Then the system should return a validation error
    And the error should indicate which field is missing
    And no complaint should be created

  Scenario: File a complaint with invalid severity
    Given I have complaint data with invalid severity
      | field         | value                     |
      |---------------|---------------------------|
      | task          | "Add user management" |
      | context        | "Current system has basic users" |
      | missing_info  | "No password reset feature" |
      | confused_by    | "UI is confusing" |
      | future_wishes  | "Better admin panel" |
      | severity       | "invalid" |
      | agent_name     | "AI Assistant" |
      | project_name   | "user-management" |
    And the MCP client is connected
    When I attempt to file the complaint
    Then the system should return a validation error
    And the error should indicate valid severity options
    And no complaint should be created

  Scenario: List complaints with no filters
    Given I have no specific filters
    And the MCP client is connected
    And there are complaints in the system
    When I list complaints using the list_complaints tool
    Then I should receive a list of all complaints
    And the response should include complaint details
    And the list should be properly formatted

  Scenario: List complaints filtered by project
    Given I have project filter "user-auth-service"
    And the MCP client is connected
    And there are multiple complaints from different projects
    When I list complaints with project filter
    Then I should receive only complaints from the specified project
    And other project complaints should be excluded
    And the response should be properly formatted

  Scenario: List complaints filtered by unresolved status
    Given I want only unresolved complaints
    And the MCP client is connected
    And there are resolved and unresolved complaints
    When I list complaints with unresolved filter
    Then I should receive only unresolved complaints
    And resolved complaints should be excluded
    And the response should show proper status

  Scenario: Resolve an existing complaint
    Given I have an existing unresolved complaint
    And I know the complaint ID
    And the MCP client is connected
    When I resolve the complaint using resolve_complaint tool
    Then the complaint should be marked as resolved
    And the resolution should persist
    And a success response should be returned

  Scenario: Resolve a non-existent complaint
    Given I have a non-existent complaint ID
    And the MCP client is connected
    When I attempt to resolve the complaint
    Then the system should return a not found error
    And the error should indicate the complaint doesn't exist
    And no complaint should be modified

  Scenario: File complaint with maximum content size
    Given I have a complaint with very large content
      | field         | value                     |
      |---------------|---------------------------|
      | task          | "A" repeated 10000 times |
      | context        | "Large amount of context" |
      | missing_info  | "Many details" |
      | confused_by    | "Too much information" |
      | future_wishes  | "Content limits" |
      | severity       | "low" |
      | agent_name     | "AI Assistant" |
      | project_name   | "stress-test" |
    And the MCP client is connected
    When I attempt to file the complaint
    Then the system should validate the content size
    And return an appropriate error if content exceeds limits
    And ensure the complaint handling remains performant

  Scenario: File complaint with special characters
    Given I have complaint data with special characters
      | field         | value                     |
      |---------------|---------------------------|
      | task          | "Test with emoji 🚀 and special chars" |
      | context        | "Contains quotes and newlines" |
      | missing_info  | "Needs unicode support" |
      | confused_by    | "Character encoding issues" |
      | future_wishes  | "Better char handling" |
      | severity       | "medium" |
      | agent_name     | "AI Assistant" |
      | project_name   | "unicode-test" |
    And the MCP client is connected
    When I attempt to file the complaint
    Then the system should handle special characters properly
    And the content should be stored correctly
    And the file should be readable
    And the response should be properly formatted

  Scenario: Concurrent complaint filing
    Given I have multiple valid complaints
    And the MCP client is connected
    When I file multiple complaints simultaneously
    Then all complaints should be stored successfully
    And each complaint should have unique IDs
    And the responses should not conflict
    And the system should handle concurrent operations gracefully

  Scenario: File complaint with automatic project detection
    Given I have complaint data without project name
      | field         | value                     |
      |---------------|---------------------------|
      | task          | "Add dark mode support" |
      | context        | "Need dark theme for night use" |
      | missing_info  | "No automatic theme detection" |
      | confused_by    | "UI hurts eyes at night" |
      | future_wishes  | "Smart theme switching" |
      | severity       | "low" |
      | agent_name     | "AI Assistant" |
      | project_name   | "" |
    And the MCP client is connected
    When I file the complaint without specifying project
    Then the system should auto-detect the project name
    And the complaint should be associated with the detected project
    And the response should include the project name
    And the detected project should be reasonable

  Scenario: File complaint and then list it back
    Given I have valid complaint data
    And the MCP client is connected
    When I file the complaint
    And then immediately list all complaints
    Then the newly filed complaint should appear in the list
    And the complaint should have all expected fields
    And the timestamp should be recent
    And the list response should be properly formatted

  Scenario: Handle network interruptions during filing
    Given I have valid complaint data
    And the MCP client connection is unstable
    When I file the complaint during network issues
    Then the system should handle the interruption gracefully
    And provide appropriate error feedback
    And attempt recovery if possible
    And maintain data consistency

  Scenario: Validate complaint with all severity levels
    Given I have complaint data for each severity level
      | severity       | description                                |
      |---------------|-------------------------------------------|
      | low            | "Minor usability issue"                    |
      | medium         | "Moderate functionality problem"             |
      | high           | "Critical functionality broken"               |
      | critical       | "Security vulnerability or data loss"          |
    And the MCP client is connected
    When I validate each severity level
    Then all severity levels should be accepted
    And invalid severity should be rejected
    And the validation should be consistent
    And appropriate error messages should be returned

  Scenario: Performance test with many complaints
    Given I have a large number of complaint data
    And the MCP client is connected
    When I file multiple complaints rapidly
    Then the system should handle the load efficiently
    And response times should remain acceptable
    And no memory leaks should occur
    And all operations should complete successfully