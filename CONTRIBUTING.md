# Contributing to sygkro

Thank you for your interest in contributing to sygkro! We welcome contributions of all kinds—from bug reports and feature requests to pull requests with code improvements. This document provides guidelines to help you get started and ensure that your contributions align with our project’s goals and coding standards.

## Table of Contents

- [Contributing to sygkro](#contributing-to-sygkro)
  - [Table of Contents](#table-of-contents)
  - [Reporting Issues](#reporting-issues)
  - [Code Contributions](#code-contributions)
    - [Getting Started](#getting-started)
    - [Branching and Commit Messages](#branching-and-commit-messages)
    - [Pull Request Guidelines](#pull-request-guidelines)
    - [Communication](#communication)
    - [License](#license)

## Reporting Issues

- **Search First:** Before opening a new issue, please search the existing issues to see if your problem or feature request has already been reported.
- **Provide Details:** When reporting a bug or suggesting a feature, include:
  - A clear description of the issue.
  - Steps to reproduce the problem (if applicable).
  - Expected and actual behavior.
  - Relevant environment details (e.g., OS, Go version, Git configuration).
- **Screenshots and Logs:** Attach screenshots, logs, or error messages to help us understand the issue.

## Code Contributions

We appreciate your help improving sygkro. Here’s how to get started:

### Getting Started

1. **Fork the Repository:**  
   Fork the sygkro repository on GitHub.

2. **Clone Your Fork Locally:**  
  
  ```bash
   git clone https://github.com/yourusername/sygkro.git
   cd sygkro
  ```

3. **Create a Feature Branch:**  
  
  ```bash
   git checkout -b feature/your-feature-name
  ```

### Branching and Commit Messages

> This project uses lefthook to perform commit and branch name validation. Please ensure you have it installed and configured correctly.

- Branch Naming:
  Use descriptive branch names, e.g., fix/issue-123 or feature/add-sync-diff.

- Commit Messages:
  Write clear, concise commit messages. We recommend following the Conventional Commits standard:
  - feat: A new feature.
  - fix: A bug fix.
  - docs: Documentation changes.
  - style: Formatting or style changes (no code change).
  - refactor: Code refactoring without feature changes.
  - test: Adding or modifying tests.

> Conventional commits are used to generate the changelog and versioning. Please follow the format strictly.
> For more information, see [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/).

- Coding Style Guidelines

- Formatting:
  Run `gofmt` on your code before submitting a pull request.

- Idiomatic Go:
  Write clear, idiomatic Go code. Follow existing patterns in the project.

- Documentation:
  Update or add documentation/comments as needed to explain your changes.

### Pull Request Guidelines

- Ensure Tests Pass:
  Run all tests with `go test ./...` before submitting your PR. This will be run automatically by the CI.

- Describe Your Changes:
  In your pull request description, include:
  - What changes you made.
  - Why these changes are necessary.
  - Any relevant issue numbers (e.g., "Fixes #123").

- Keep It Focused:
  Try to keep each pull request focused on a single change or feature.

### Communication

- Discuss Before Large Changes:
  If you’re planning significant changes or new features, open an issue or join our discussion forum to gather feedback before starting.

- Be Respectful and Constructive:
  We value a collaborative and welcoming community. Please be respectful in all interactions.

### License

By contributing to sygkro, you agree that your contributions will be licensed under the MIT License.