# sygkro

> This project is in early development

**sygkro** is a project templating and synchronization tool written in Go inspired by [Cookiecutter](https://github.com/cookiecutter/cookiecutter) and [Cruft](https://github.com/cruft/cruft). It helps you create, manage, and update projects based on customizable templates, keeping your projects in sync with evolving boilerplate over time.

## Table of Contents

- [sygkro](#sygkro)
  - [Table of Contents](#table-of-contents)
  - [Features](#features)
  - [Installation](#installation)
    - [Prerequisites](#prerequisites)
    - [Building from Source](#building-from-source)
    - [Usage](#usage)
      - [Creating a New Template](#creating-a-new-template)
      - [Creating a New Project](#creating-a-new-project)
      - [Viewing Differences](#viewing-differences)
      - [Applying Diffs](#applying-diffs)
    - [Configuration Files](#configuration-files)
    - [Git \& Diff Integration](#git--diff-integration)
    - [Contributing](#contributing)
    - [License](#license)

## Features

- **Project Templating:**  
  Create new project templates with default configurations and sample files.
- **Project Generation:**  
  Generate new projects from templates (local or Git-based) using customizable inputs.
- **Diff & Sync:**  
  Compare your project against its originating template with a Git-style diff.
- **Git Integration:**  
  Clone template repositories via SSH or HTTPS, using simplified syntax (e.g. `gh:owner/repo`), and track template versions using commit SHAs.

## Installation

### Prerequisites

- [Go](https://golang.org/) 1.18+
- Git (configured for SSH and/or HTTPS)

### Building from Source

Clone the repository:

```bash
git clone https://github.com/faraday/sygkro.git
cd sygkro
```

Build the binary:

```bash
go build -o sygkro ./main.go
```

Optionally, move the sygkro binary into your PATH.

### Usage

sygkro is organized into several subcommands for managing templates and projects.

#### Creating a New Template

Generate a new project template directory with default configuration:

sygkro template new `<template-name>`

This command creates a directory named `<template-name>` containing:

- A default `.sygkro.template.yaml` configuration file.
- A `README` with templating examples.

#### Creating a New Project

Generate a new project from an existing template:

```bash
sygkro project create --template <template-ref> --target <target-directory>
```

- `<template-ref>`:
  A local path or Git repository reference. Supported formats:
  - Simplified syntax: `gh:owner/repo`
  - HTTPS URL: `https://github.com/owner/repo.git`
  - SSH URL: `git@github.com:owner/repo.git`

- `<target-directory>`:
  The directory where the project will be created (defaults to the current directory).

#### Viewing Differences

To compare your project with its original template (based on the metadata in `.sygkro.sync.yaml`):

```bash
sygkro project diff
```

This command:

1. Loads the sync metadata from `.sygkro.sync.yaml`.
2. Clones the template repository at the tracked commit.
3. Re-renders the template using the stored inputs into a temporary “ideal” state.
4. Computes a unified diff between the ideal output and the current project.

#### Applying Diffs

This functionality has not been implemented yet, but you can use the `sygkro project diff` command to generate a diff output. The output can be piped into Git for applying changes. This process requires manual edits to the diff output and instructions will not be provided here as its not fully proven out.

### Configuration Files

- Template Configuration:
  Stored as `.sygkro.template.yaml` in a template directory. Defines the schema for template inputs and options (e.g. files to skip rendering).

- Sync Metadata:
  Generated projects include a `.sygkro.sync.yaml` file that stores:
  - Source: The original template reference and tracking commit SHA.
  - Inputs: The values used when generating the project.
  - Options: Additional options affecting diff/sync behavior.

### Git & Diff Integration

sygkro leverages Git to:

- Clone template repositories:
  Supports SSH and HTTPS URLs, as well as a simplified gh: syntax.

- Version Tracking:
  Uses the HEAD commit SHA from the cloned template as the template version in the sync metadata.

- Computing Diffs:
  Re-renders the template into a temporary “ideal” state and computes a unified diff between that state and your current project directory.

### Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on contributing to sygkro.

### License

This project is licensed under the MIT License. See [LICENSE.md](LICENSE.md) for details.