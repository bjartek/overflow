# Contributing to the Flow CLI

The following is a set of guidelines for contributing to Overflow
These are mostly guidelines, not rules.
Use your best judgment, and feel free to propose changes to this document in a pull request.
These guidelines are loosely based on the guidelines found in flow-cli

## Table Of Contents

[Getting Started](#project-overview)

[How Can I Contribute?](#how-can-i-contribute)

- [Reporting Bugs](#reporting-bugs)
- [Suggesting Enhancements](#suggesting-enhancements)
- [Pull Requests](#pull-requests)

[Styleguides](#styleguides)

- [Git Commit Messages](#git-commit-messages)
- [Go Styleguide](#go-styleguide)

[Additional Notes](#additional-notes)

## How Can I Contribute?

### Reporting Bugs

#### Before Submitting A Bug Report

- **Search existing issues** to see if the problem has already been reported.
  If it has **and the issue is still open**, add a comment to the existing issue instead of opening a new one.

#### How Do I Submit A (Good) Bug Report?

Explain the problem and include additional details to help maintainers reproduce the problem:

- **Use a clear and descriptive title** for the issue to identify the problem.
- **Describe the exact steps which reproduce the problem** in as many details as possible.
  When listing steps, **don't just say what you did, but explain how you did it**.
- **Provide specific examples to demonstrate the steps**.
  Include links to files or GitHub projects, or copy/pasteable snippets, which you use in those examples.
  If you're providing snippets in the issue,
  use [Markdown code blocks](https://help.github.com/articles/markdown-basics/#multiple-lines).
- **Describe the behavior you observed after following the steps** and point out what exactly is the problem with that behavior.
- **Explain which behavior you expected to see instead and why.**
- **Include error messages and stack traces** which show the output / crash and clearly demonstrate the problem.

Provide more context by answering these questions:

- **Can you reliably reproduce the issue?** If not, provide details about how often the problem happens
  and under which conditions it normally happens.

Include details about your configuration and environment:

- **What is the version of the CLI you're using**?
- **What's the name and version of the Operating System you're using**?

### Suggesting Enhancements

#### Before Submitting An Enhancement Suggestion

- **Perform a cursory search** to see if the enhancement has already been suggested.
  If it has, add a comment to the existing issue instead of opening a new one.

#### How Do I Submit A (Good) Enhancement Suggestion?

Enhancement suggestions are tracked as [GitHub issues](https://guides.github.com/features/issues/).
Create an issue and provide the following information:

- **Use a clear and descriptive title** for the issue to identify the suggestion.
- **Provide a step-by-step description of the suggested enhancement** in as many details as possible.
- **Provide specific examples to demonstrate the steps**.
  Include copy/pasteable snippets which you use in those examples,
  as [Markdown code blocks](https://help.github.com/articles/markdown-basics/#multiple-lines).
- **Describe the current behavior** and **explain which behavior you expected to see instead** and why.
- **Explain why this enhancement would be useful** to overflow users.


### Pull Requests

The process described here has several goals:

- Maintain code quality
- Fix problems that are important to users
- Engage the community in working toward the best possible UX
- Enable a sustainable system for the maintainers to review contributions

Please follow the [styleguides](#styleguides) to have your contribution considered by the maintainers.
Reviewer(s) may ask you to complete additional design work, tests,
or other changes before your pull request can be ultimately accepted.

## Styleguides

Before contributing, make sure to examine the project to get familiar with the patterns and style already being used.

### Git Commit Messages

- Use the present tense ("Add feature" not "Added feature")
- Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
- Limit the first line to 72 characters or less
- Reference issues and pull requests liberally after the first line

### Go Styleguide

The majority of this project is written Go.

We try to follow the coding guidelines from the Go community.

- Code should be formatted using `gofmt`
- Code should pass the linter: `make lint`
- Code should follow the guidelines covered in
  [Effective Go](http://golang.org/doc/effective_go.html)
  and [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Code should be commented
- Code should pass all tests: `make test`

## Inspiration and Reference
https://clig.dev/

https://blog.developer.atlassian.com/10-design-principles-for-delightful-clis/

https://devcenter.heroku.com/articles/cli-style-guide

https://eng.localytics.com/exploring-cli-best-practices/

