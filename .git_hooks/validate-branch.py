#!/usr/bin/env python3
import sys
import re
import subprocess


def get_current_branch():
    """Retrieve the current Git branch name."""
    try:
        branch = subprocess.check_output(["git", "rev-parse", "--abbrev-ref", "HEAD"], encoding="utf-8").strip()
        return branch
    except subprocess.CalledProcessError:
        sys.stderr.write("Error: Could not determine the current branch.\n")
        sys.exit(1)


def main():
    # Use branch name from command-line argument if provided; otherwise, get current branch.

    branch_name = get_current_branch()

    # Define the allowed trunk branch names.
    trunk_branches = {"main", "master", "trunk"}

    if branch_name in trunk_branches:
        print(f"Branch '{branch_name}' is a trunk branch. (Allowed)")
        sys.exit(0)

    pattern = r"^(cruft|feature|bugfix|hotfix|release)/([a-z0-9\-]+)?$"

    if re.match(pattern, branch_name):
        print(f"Branch '{branch_name}' is valid for trunk based development.")
        sys.exit(0)
    else:
        sys.stderr.write(f"Error: Branch name '{branch_name}' does not conform to the required naming convention.\n")
        sys.stderr.write("Valid branch names are:\n")
        sys.stderr.write("  - Trunk branches: main, master, or trunk\n")
        sys.stderr.write("  - Bugfix branches: bugfix/description\n")
        sys.stderr.write("  - Cruft branches: cruft/update\n")
        sys.stderr.write("  - Feature branches: feature/description\n")
        sys.stderr.write("  - Hotfix branches: hotfix/description\n")
        sys.stderr.write("  - Release branches: release/description\n")
        sys.exit(1)


if __name__ == "__main__":
    main()
