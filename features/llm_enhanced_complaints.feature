Feature: LLM-Enhanced Complaint Filing
  As an AI coding assistant with advanced MCP capabilities
  I want to file complaints with AI assistance and structured guidance
  So that I can provide higher quality feedback and better organized information

  Background:
    Given a temporary project directory is set up
    And complaints-mcp server is started with advanced features

  Scenario: Use AI assistance to generate complaint content
    When I request AI assistance for filing a complaint about:
      | issue_type | confusing documentation |
      | context   | missing API endpoints |
    And I provide the raw information: "The setup instructions mention API endpoints but don't specify what they are"
    Then the AI should suggest structured complaint fields
    And the suggested complaint should be properly formatted
    And the AI should generate appropriate future wishes

  Scenario: File complaint with progress tracking
    When I start a complex complaint filing process
    And the process requires multiple steps to complete
    Then I should receive progress notifications
    And the progress should update from 0% to 100%
    And the final complaint should be saved when progress reaches 100%

  Scenario: Access existing complaints as MCP resources
    When I list available MCP resources
    Then previously filed complaints should be available as resources
    And I should be able to read a specific complaint by URI
    And the complaint resource should return markdown content
    And the resource should support text/plain MIME type

  Scenario: Use structured prompts for complaint guidance
    When I access the complaint-guidance prompt
    And I specify: "I'm confused about the project structure"
    Then the prompt should return structured guidance questions
    And the guidance should help me identify missing information
    And the guidance should suggest relevant fields to fill out

  Scenario: Validate tool schemas automatically generated
    When I inspect the file_complaint tool schema
    Then the schema should include all required field definitions
    And each field should have proper JSON schema types
    And field descriptions should be helpful for AI agents
    And optional fields should be marked as non-required

  Scenario: Handle concurrent complaint filing safely
    When I file multiple complaints simultaneously
    And each complaint has different session names
    Then the server should handle all requests without conflicts
    And each complaint should be saved with unique filenames
    And the global storage should organize complaints by project

  Scenario: Use pagination for large complaint sets
    When I file 50 complaints in quick succession
    And I request to list complaints with pagination
    Then I should receive complaints in pages of 10 items each
    And I should be able to navigate through all pages
    And the total count should be accurate

  Scenario: Test error handling and validation
    When I file a complaint with invalid JSON data
    Then the server should return a validation error
    And the error message should indicate the specific validation issue
    When I provide a negative number in a numeric field
    Then the server should reject the input with a type error
    When I exceed the maximum allowed complaint length
    Then the server should return an appropriate size limit error

  Scenario: Integrate with client-side roots management
    Given I have configured client roots for my project directories
    When I connect to the complaints MCP server
    Then the server should be notified of my configured roots
    And the server should be able to access files in my configured directories
    And I should be able to file complaints that reference files in my roots

  Scenario: Test middleware chain execution
    Given the server has logging middleware configured
    And the server has authentication middleware configured
    And the server has validation middleware configured
    When I file a valid complaint
    Then the request should pass through all middleware layers
    And the logging should capture the request details
    And the authentication should validate the request
    And the validation should ensure data integrity

  Scenario: Use sampling for intelligent complaint analysis
    When I have a complex situation requiring analysis
    And I request AI sampling to analyze my situation
    Then the AI should provide contextual analysis
    And the analysis should identify the core issues
    And the analysis should suggest specific actionable improvements
    And the complaint should include both my input and AI analysis