Feature: Complaint Filing
  As an AI coding assistant
  I want to file complaint reports about missing or confusing information
  So that I can help improve the project documentation and developer experience

  Background:
    Given a temporary project directory is set up
    And the complaints-mcp server binary is available

  Scenario: File a complete complaint with all required fields
    When I file a complaint with:
      | task_asked_to_perform    | Implement a REST API |
      | context_information       | Only basic requirements given |
      | missing_information       | No API endpoint specifications |
      | confused_by             | Missing database schema information |
      | future_wishes           | Complete API documentation |
      | session_name            | api-session |
      | agent_name              | Claude AI |
    Then the complaint should be saved locally
    And the complaint should be saved globally with project name
    And the complaint content should contain all provided information
    And the filename should follow the expected format

  Scenario: File a complaint with minimal required fields
    When I file a complaint with:
      | task_asked_to_perform    | Setup development environment |
      | context_information       | GitHub repository cloned |
      | missing_information       | Environment variables not documented |
      | confused_by             | Configuration unclear |
      | future_wishes           | Better setup instructions |
    Then the complaint should be saved locally
    And the complaint should be saved globally with project name
    And default values should be used for optional fields

  Scenario: File a complaint in a non-git repository
    Given a non-git repository directory
    When I file a complaint with:
      | task_asked_to_perform    | Test project setup |
      | context_information       | Basic folder structure |
      | missing_information       | Missing build instructions |
      | confused_by             | Project organization unclear |
      | future_wishes           | Clear documentation |
    Then the complaint should be saved locally
    And the global complaint should use folder name as project name

  Scenario: Handle file system errors gracefully
    Given a read-only project directory
    When I attempt to file a complaint
    Then the complaint filing should fail with an appropriate error message

  Scenario: Generate consistent complaint filenames
    When I file multiple complaints with different session names
    Then each filename should be unique
    And each filename should follow the timestamp-session format

  Scenario: Validate complaint content structure
    When I file a complete complaint
    Then the complaint markdown should have the correct headers
    And the complaint markdown should include the current date
    And the complaint markdown should include agent name
    And the complaint markdown should be properly formatted