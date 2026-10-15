# Architecture decision records

ADRs record decisions that are expensive to reverse or that define a product
guarantee. They supplement, rather than replace, the current architecture in
the parent documentation.

Status values are **Proposed**, **Accepted**, **Superseded**, or **Deprecated**.
A later ADR supersedes an earlier one; accepted records are not rewritten to
hide old trade-offs.

- [0001 — Defer creation of missing command-line files](0001-deferred-file-creation.md)
- [0002 — Put the discoverable command menu at the bottom](0002-bottom-command-menu.md)

Earlier foundational choices—Go, tcell, a line-slice buffer, Chroma, and
anchor/head selections—are documented in
[ARCHITECTURE.md](../ARCHITECTURE.md). They predate this ADR series.
