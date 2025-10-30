Feature: Project Name Detection
  As the complaints-mcp server
  I want to automatically detect project names from git repositories
  So that complaints can be organized by project in global storage

  Background:
    Given a temporary project directory is set up

  Scenario: Detect project name from git remote
    Given a git repository with remote "origin" pointing to "https://github.com/user/project-name.git"
    When the project name is detected
    Then the project name should be "project-name"

  Scenario: Detect project name from git remote with .git suffix
    Given a git repository with remote "origin" pointing to "git@github.com:user/another-project.git"
    When the project name is detected
    Then the project name should be "another-project"

  Scenario: Fallback to folder name when no git remote exists
    Given a non-git directory named "my-awesome-project"
    When the project name is detected
    Then the project name should be "my-awesome-project"

  Scenario: Fallback to default when no suitable name found
    Given a directory in root with no git setup
    When the project name is detected
    Then the project name should be "unknown-project"

  Scenario: Handle git command failures gracefully
    Given a directory where git commands fail
    When the project name is detected
    Then the project name should fallback to folder name