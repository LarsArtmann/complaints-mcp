Feature: MCP Server Integration
  As an MCP client
  I want to connect to the complaints-mcp server
  So that AI agents can file complaints through the MCP protocol

  Background:
    Given the complaints-mcp server is started
    And a client connection is established

  Scenario: List available tools
    When I request the list of tools
    Then the "file_complaint" tool should be available
    And the tool should have the correct description
    And the tool should have all required input schema properties

  Scenario: File complaint through MCP tool
    When I call the "file_complaint" tool with valid arguments:
      | task_asked_to_perform    | Test MCP integration |
      | context_information       | Server is running |
      | missing_information       | None - this is a test |
      | confused_by             | No issues |
      | future_wishes           | Everything works perfectly |
    Then the tool should return a success message
    And the complaint should be saved to the expected locations

  Scenario: Handle invalid tool arguments
    When I call the "file_complaint" tool with missing required arguments
    Then the tool should return an error
    And the error message should indicate the missing fields

  Scenario: Tool validation and schema
    When I inspect the "file_complaint" tool schema
    Then all required fields should be defined
    And optional fields should be marked appropriately
    And field descriptions should be helpful