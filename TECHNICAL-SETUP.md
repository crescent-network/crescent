# Technical Setup

## Setup git hooks for Conventional Commit

1. Install [`pre-commit`](https://pre-commit.com/)
2. Run this command:

```
$ pre-commit install --hook-type commit-msg
```
  
3. (Optional for macOS users) Install GNU `grep`:
  1. Run `brew install grep`
  2. Add this line to your shell profile:

```
export PATH="/usr/local/opt/grep/libexec/gnubin:$PATH"
```

Now, whenever you make a commit, the `pre-commit` hook will be run to check if the commit message
conforms [Conventional Commit](https://www.conventionalcommits.org/) rule.

