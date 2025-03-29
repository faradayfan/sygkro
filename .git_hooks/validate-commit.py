#!/usr/bin/env python3
import sys
import re


def main():
    if len(sys.argv) < 2:
        sys.stderr.write("Error: Commit message file not provided.\n")
        sys.exit(1)

    commit_msg_file = sys.argv[1]
    try:
        with open(commit_msg_file, "r", encoding="utf-8") as f:
            commit_message = f.read().strip()
    except Exception as e:
        sys.stderr.write(f"Error reading commit message file: {e}\n")
        sys.exit(1)

    # Regular expression for Conventional Commits:
    # Matches a commit message that starts with a valid type (feat, fix, docs, etc.),
    # an optional scope in parentheses, an optional exclamation mark for breaking changes,
    # a colon, a space, and then the description.
    pattern = r"^(feat|fix|docs|style|refactor|perf|test|chore|ci|build)(\([\w\-.]+\))?(!)?: .+"
    if not re.match(pattern, commit_message):
        sys.stderr.write("\n✖  Invalid commit message format.\n\n")
        sys.stderr.write("Please follow the Conventional Commits format. Examples:\n")
        sys.stderr.write("  feat(parser): add ability to parse arrays\n")
        sys.stderr.write("  fix: correct minor typos in code\n\n")
        sys.exit(1)

    print("✓  Commit message is valid.")
    sys.exit(0)


if __name__ == "__main__":
    main()
