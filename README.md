# gh-dispatch

`gh-dispatch` is an extension to the [gh CLI](https://cli.github.com/) for triggering [repository_dispatch](https://docs.github.com/en/rest/repos/repos#create-a-repository-dispatch-event) and
[workflow_dispatch](https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#workflow_dispatch) events and watching the resulting GitHub Actions workflow run.

## Installation

Install the `gh` CLI [for your platform](https://github.com/cli/cli#installation). For example, on Mac OS:

```
brew install gh
```

Install the latest `dispatch` extension from [its pre-compiled releases](https://github.com/mdb/gh-dispatch/releases):

```
gh extension install mdb/gh-dispatch
```

## Development

Build and test `gh-dispatch` locally:

```
make
```

Install a locally built `gh-dispatch`, thus making it available `gh dispatch`:

```
make install
```
