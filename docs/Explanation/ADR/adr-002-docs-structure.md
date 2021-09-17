# ADR 002: Documentation Structure

## Status

ACCEPTED

## Abstract

This ADR proposes a documentation strategy based on the *Grand Unified Theory of Documentation* (David Laing) as described by [Divio](https://documentation.divio.com/).

The documentation strategy outlines four specific use cases for documentation. Based on these use cases and other non-functional requirements, a structure is proposed that will address these concerns using GitHub as the Content Management System. 

The documentation strategy also proposes:

- The use and re-use of document and format templates
- Specific [code owners](https://docs.github.com/en/github/creating-cloning-and-archiving-repositories/creating-a-repository-on-github/about-code-owners#about-code-owners) for documentation 
- Comment and commit templates combined with [githook](https://git-scm.com/docs/githooks) checks

The outcome shall be focused, consistent, high quality documentation. 

## Context

Good documentation is important to the success of software projects.

*Writing excellent code doesn't end when your code compiles or even if your test coverage reaches 100%. It's easy to write something a computer understands, it's much harder to write something both a human and a computer understand. Your mission as a Code Health-conscious engineer is to write for humans first, computers second. Documentation is an important part of this skill.* [Google Documentation Best Practice](https://google.github.io/styleguide/docguide/best_practices.html)

The documentation use cases, as outlined by Divio are:

- Allow a new user to get started
- Show a user how to solve a specific problem
- Describe the machinery, for example, classes, functions, interfaces, parameters, and so on
- Explanation and context for design, scope, and so on

![Documentation Quadrants](https://documentation.divio.com/_images/overview.png) 

The goals of well-structured and well-written documentation include:

- Findability: depending on the use case, the technical content can be discovered and accessed
- Style: The documentation is written in an appropriate style for the use case
- Consistency: Each type of documentation is written in a consistent style
- Scoped: Documentation is scoped to a specific use case; for example, a tutorial can provide links but does not include technical content that describes why the software works, a tutorial just teaches how to use it

Additional Documentation non-functional use cases include:

- Technical content SHOULD BE as close to the code as reasonably practicable and strive to use the docs as code workflow
- Technical content SHOULD BE generated from code as much as possible
- Technical content SHOULD USE a consistent format 
- Technical content SHOULD BE useable from within the repository
- Technical content COULD HAVE an automatic process that converts the content to a website based on [Read The Docs](https://readthedocs.com/), [Gitbook](https://www.gitbook.com/), or other suitable hosting systems

## Decision

To address the use cases outlined in the context, this ADR proposes the following decisions:

- Use GitHub as primary content management [https://github.com/tendermint/farming](https://github.com/tendermint/farming)
- Use Markdown and LaTeX to deliver research publications

Given GitHub will form the content management system, we propose the following structure:

### Structure

The documentation structure shall use as much as possible a content structure similar to the [Divio user cases](https://documentation.divio.com/introduction/).

|                 | Tutorials | How-to guides | Reference   | Explanation   |
|-----------------|-----------|---------------|-------------|-------------- |
| **Oriented to** | Learning  | A goal        | Information | Understanding | 
| **Must**        | Allow a newcomer to get started | Show how to solve a specific problem | Describe the machinery | Explain |
| **Takes the form of**    | A lesson | A series of steps | A dry description | A discursive explanation |
| **Analogy**     | Teaching a child to cook | Recipe in a cookery book | An encyclopedia article | A paper on culinary social history |

The specific implementation for farming module SHOULD BE as per the following tree structure.

```
/
├── README
├── CONTRIBUTING
├── TECHNICAL-SETUP
├── CODEOWNERS
├── x/
|   ├── module_a/
|       ├── README
|       ├── docs/
|           ├── state
|           ├── state_transitions
|           ├── messages
├── docs/
    ├── README
    ├── CODEOWNERS
    ├── Explanation/
    |   ├── README
    |   ├── ADR/
    |   |   ├── README
    |   |   ├── PROCESS
    |   |   ├── adr-template
    |   |   ├── adr-{number}-{desc}
    |   ├── articles/
    |   |   ├── regulation-litepaper/
    |   |       ├── ARTICLE
    |   ├── research/
    |       ├── README
    |       ├── research_topic/
    ├── How-To/
    |   ├── HowToDoSomething/
    |   ├── HowToDoSomethingElse/
    ├── Reference/
    |   ├── README
    |   ├── GLOSSARY
    |   ├── MODULES
    |   ├── use-cases/
    |   |   ├── use-case-A
    |   |   ├── use-case-B
    |   ├── architecture/
    ├── Tutorials/
        ├── Tutorial_1/
        ├── Tutorial_2/
```

#### Root level documents

The following files are required at the repo root level:

- **README.md** - General repo overview to introduce the product and orientate the user. All README files must follow the best practices as outlined in the [GitHub README](https://docs.github.com/en/github/creating-cloning-and-archiving-repositories/creating-a-repository-on-github/about-readmes) guidelines.
- **TECHNICAL-SETUP.md** - Specific steps on getting started with the repo, can be a link to a tutorial or include the specific action-oriented steps
    - Links to specific tooling setup requirements for development tools, linters, and so on
    - Dependencies such as [pre-commit](https://pre-commit.com/) package manager
    - Building the code
    - Running tests
- **CONTRIBUTING.md** - Details on how new users can contribute to the project. In specific:
    - Committing changes
    - Commit message formats (see [Commit Comments](#commit-comments)
    - Raising PRs
    - Code of Conduct
- **CODEOWNERS** - Although not part of the documentation itself, a [CODEOWNERS file](https://docs.github.com/en/github/creating-cloning-and-archiving-repositories/creating-a-repository-on-github/about-code-owners) defines the code maintainers who are responsible for code in a repository and perform quality assurance on comments, PRs, and issues.

#### Modules

In line with Cosmos SDK convention (TODO: needs reference) each module contains its own relevant documentation:

- **Module specifications** - A document that outlines state transitions `x/module-name/docs/`
- **Module-level README.md** e.g. x/module-name/README.md

README files are classed as reference documentation. Content in module-level README files is descriptive, but explanatory. Explanations should be part of issues, Pull Requests, and docs/explanation/architecture.

#### docs/

The `docs` folder shall include the following files and folders:

- **README.md** - SHALL USE this for introduction and orientating the user, based on the content of this ADR and other materials.
- **CODEOWNERS** - This [CODEOWNERS file](https://docs.github.com/en/github/creating-cloning-and-archiving-repositories/creating-a-repository-on-github/about-code-owners) details the reviewers for documentation folder. The listed code owners SHALL INCLUDE the code maintainers in the root CODEOWNERS file plus a member of the Tendermint Technical Writing Team.

#### docs/Reference

Reference documentation includes a number of different forms:

- **README.md** - This document outlines the purpose of the reference documentation as per the use-case documentation strategy and methodology. In addition, the README also links to documentation that is created from the code itself, specifically:
    - Code Documentation in form of Go Docs
    - Swagger API documentation
- **GLOSSARY.md** - Review and maintenance must be regularly and consistently applied. These form the terms of reference for users and ensure that discussion and design are based on consistent terms of reference. This file will be similar to [Cosmos Network Glossary](https://v1.cosmos.network/glossary) and can reference this.
- **MODULES.md** - A markdown document that has references to module-relevant documentation

##### docs/Reference/use-cases

The `use-cases` folder describes the farming module use cases. Ideally, use cases are written in behavior-driven development (BDD) format. Use case content should be dry in nature and avoid explanations that should be covered in the explanation documentation.

##### docs/Reference/architecture

The `architecture` folder contains architecture diagrams such as component, activity, and sequence diagrams as relevant. Specifically, these assets should be in a format suitable for version management and easy to update. Therefore, these diagrams should be in SVG or DOT format and not image formats (JPEG, PNG, and so on).

#### docs/Explanation

The `Explanation` folder contains content that provides context for readers and is discursive in nature. See the [Divio Explanation page](https://documentation.divio.com/explanation/#) for more detail.

- **docs/explanation/README.md** - This file orients the reader and explains the content. 

##### docs/Explanation/ADR

The `ADR` folder tracks decisions regarding design and architecture (such as this documentation strategy). ADR content includes the following:

- **docs/explanation/adr/README** - introduction to ADR
- **docs/explanation/adr/PROCESS.md** - describes how to raise ADRs
- **docs/explanation/adr/adr-template.md** - template for raising ADR
- **docs/explanation/adr/adr-{number}-{desc}.md** - an ADR document

##### docs/Explanation/articles

The `articles` folder contains a sub-folder for each published article. Published articles this COULD REFER to blog posts. The folder should be named such that it describes the article's purpose. Each sub-folder SHALL CONTAIN all the content relevant to the article (for example, images, bibliographies, and so on). These articles can be converted into PDF format using Pandoc. 

To convert articles to PDF using Pandoc:

- There SHOULD BE a makefile with targets for calling Pandoc. Note: the process for building PDF files is not part of the commit or release processes, but ad-hoc
- There SHOULD BE a LaTeX template file that can create PDF files that have a consistent look and feel. This COULD BE the [Eisvogel template](https://github.com/Wandmalfarbe/pandoc-latex-template) with suitable modifications.
- The makefile and template should be independent of the article
- There SHOULD BE a README.md that describes how to use the makefile and template and build articles

> **Note:** Explanations can come in other forms, particularly issue discussion and Pull Requests.

#### docs/Tutorials

As indicated in the overview, tutorials SHALL BE documents that target beginners and guide a user step-by-step through a process with the aim of achieving some goal. Please see the [Divio tutorial page](https://documentation.divio.com/tutorials/) for details.

- There SHALL BE a folder for each tutorial. See the [Cosmos SDK tutorials](https://github.com/cosmos/sdk-tutorials) as an example.
- The folder SHALL CONTAIN all of the content that is relevant for that tutorial. 
- The content SHOULD BE consistent in format with [Cosmos SDK tutorials](https://tutorials.cosmos.network/). 

#### docs/How-To

In contrast to tutorials, [how-to guides](https://documentation.divio.com/how-to-guides/) are a series of actionable steps to help an experienced reader solve a specific problem. These how-to guides SHALL USE templates similar to the tutorials - see above.

### Templates

The documentation SHOULD USE Markdown templates to develop structured technical content, like module messages follow templates in the Cosmos SDK. 

- [The good docs project](https://github.com/thegooddocsproject)
- [Readme editor](https://readme.so/editor)


#### Code Comments

PR review comments also form part of the documentation. Comments SHALL FOLLOW recommendation as per [Conventional Comments](https://conventionalcomments.org/)

```
<label> [decoration]: <subject>

[discussion]
```

where `label = (praise|nitpick|suggestion|issue|question|thought|chore)`

#### Commit Comments

Commits comments will also follow a similar format as laid out following the standard defined by [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/#summary). This commit convention SHOULD BE enforced as part of [pre-commit](https://pre-commit.com/) checks.

## Consequences

This section describes the resulting context, after applying the decision. 

### Backwards Compatibility

After this ADR is implemented, existing documentation will be migrated from existing sources that include:

- Notion
- Other Git repos
- Published papers
- Blog posts 

### Positive

As a result of this documentation strategy:

- Content development and maintenance will follow best practices that ensure content is easy to navigate and read
- Content will be in a consistent format
- Commits, Issues, and Pull Requests in the repo will follow best practices
- CHANGELOG and release documentation will benefit from better commit messages, reducing developer effort

### Negative

- There may be more effort required
- Moving modules into new repos may cause inconsistenties in the repo

## Further Discussions

While an ADR is in the DRAFT or PROPOSED stage, this section should contain a summary of issues to be solved in future iterations (usually referencing comments from a pull-request discussion).

Later, this section can optionally list ideas or improvements the author or reviewers found during the analysis of this ADR.

## References

- [Google Style Guide for Markdown](https://github.com/google/styleguide/blob/gh-pages/docguide/style.md)
- [Write the Docs global community](https://www.writethedocs.org/)
- [Write the Docs Code of Conduct](https://www.writethedocs.org/code-of-conduct/#the-principles)