# ADR 0001: Defer creation of missing command-line files

- Status: Accepted
- Date: 2026-10-15
- Release: 1.1.0

## Context

Users expect `axam notes1.txt` to behave like a mainstream editor: open a
named empty document now, but create the filesystem object only after an
explicit Save. Creating the file at startup leaves unwanted zero-byte files
when the user immediately closes the editor. Treating every missing path as a
generic untitled document loses the intended save destination.

The 1.0 implementation already opened a missing path in memory, but the rule
was implicit and accepted paths whose parent directory did not exist. The
condition was also harder to read than the product guarantee it represented.

## Decision

When an argument path does not exist:

1. Resolve it to an absolute path using the process working directory.
2. Require its parent to exist and be a directory.
3. Open an empty document with that path, `Exists=false`, and syntax selected
   from the intended filename.
4. Perform no filesystem write during open or unchanged close.
5. On explicit Save, use the normal sibling-temp atomic-save path and set
   `Exists=true` only after success.
6. Allow explicit Save of an unchanged pending document to create an empty
   file. Save is the user’s creation intent.

## Alternatives considered

### Create the file during startup

Rejected because a read/navigation action would have a persistent side
effect and failed experiments would litter directories.

### Keep the path only as display text

Rejected because Save would unexpectedly ask for a filename that the user
already supplied.

### Create missing parent directories automatically

Rejected because a typo could create an unexpected directory tree. Directory
creation is a separate, explicit operation outside the editor’s open command.

### Refuse all nonexistent arguments

Rejected because creating a named file from the command line is a basic editor
workflow.

## Consequences

Positive:

- Opening and closing an empty named document is side-effect free.
- Save has a deterministic target and remains atomic on Unix.
- Syntax highlighting can use the intended extension before the file exists.
- Errors for missing parents occur before the terminal session begins when
  invoked from the CLI.

Costs and risks:

- A pending document must be visibly distinguishable through status text and
  `Exists=false`.
- Reopen and Save As paths must not mark a document as existing before a
  successful write.
- Another process can create the path while it is pending. The first Save uses
  current save semantics; future work may add a create-only conflict check.

## Verification

Automated app tests assert:

- a relative missing filename resolves under the working directory;
- opening and unchanged close leave no file;
- editing and saving creates the expected bytes;
- explicit Save creates a zero-byte file;
- a missing parent returns an error and opens no document.

The manual release check runs `axam notes1.txt`, exits unchanged, and verifies
that `notes1.txt` does not exist.
