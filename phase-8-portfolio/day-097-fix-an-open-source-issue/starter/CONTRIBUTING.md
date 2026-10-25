# Contributing to pesa

Thanks for helping. pesa is small on purpose, so we keep changes small too.

## Before you start

- For a bug, check there's an open issue. If there isn't, open one first with the input, what you got and what you expected.
- For a new feature, open an issue and wait for a maintainer to agree before writing code. We say no to most features, to keep the package small.
- Comment on the issue to say you're working on it, so two people don't do the same fix.

## Making the change

- One fix or feature per pull request. Don't reformat or rename code you aren't fixing.
- Every bug fix comes with a test that fails before the fix and passes after it.
- Run these before you push. CI runs the same:

  ```bash
  gofmt -l .        # must print nothing
  go vet ./...
  go test -race ./...
  ```

- Don't add dependencies. pesa uses the standard library only.
- Keep the public API the same. A change to an exported name needs a major version.

## The pull request

- Fill in the template in full. Link the issue with `Fixes #NN` so it closes on merge.
- Commit messages: `pesa: <what changed>`, in the present tense, under 72 characters.
- Expect review comments. Push new commits to answer them; we squash on merge.

Be kind. We follow the Go Community Code of Conduct.
