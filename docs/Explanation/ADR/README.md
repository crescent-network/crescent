# Architecture Decision Records (ADR)

Use this location to record all high-level architecture decisions in the farming module.

### Definitions

Within the context of an ADR we define the following:
An Architectural Decision (**AD**) is a software design choice that addresses a functional or non-functional requirement that is architecturally significant.
An Architecturally Significant Requirement (**ASR**) is a requirement that has a measurable effect on a software system’s architecture and quality.
An Architectural Decision Record (**ADR**) captures a single AD, and is as often done when writing personal notes or meeting minutes. The collection of ADRs created and maintained in a project constitute its decision log. All these records are within the topic of Architectural Knowledge Management (AKM).

You can read more about the ADR concept in the [Documenting architecture decisions, the Reverb way](https://product.reverb.com/documenting-architecture-decisions-the-reverb-way-a3563bb24bd0#.78xhdix6t) blog post.

## Rationale

ADRs are intended to be the primary mechanism for proposing new feature designs and new processes, for collecting community input on an issue, and for documenting the design decisions.
An ADR should provide:

- Context on the relevant goals and the current state
- Proposed changes to achieve the goals
- Summary of pros and cons
- References
- Changelog

Note the distinction between an ADR and a specification. The ADR provides the context, intuition, reasoning, and justification for a change in architecture, or for the architecture of something new. The specification is a summary of everything as it stands today.

If recorded decisions turned out to be lacking the required substance, the process is to convene a discussion, record the new decisions here, and then modify the code to match.

## Creating new ADR

Read about the [PROCESS](./PROCESS.md).

#### Use RFC 2119 Keywords

When writing ADRs, follow the same best practices for writing RFCs. When writing RFCs, key words are used to signify the requirements in the specification. These words are often capitalized: "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL. They are to be interpreted as described in [RFC 2119](https://datatracker.ietf.org/doc/html/rfc2119).

## ADR Table of Contents

- [ADR 002: Documentation Structure](./adr-002-docs-structure.md)