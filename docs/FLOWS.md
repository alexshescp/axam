# Runtime and file-safety flows

These diagrams describe axam 1.1 behavior. They are architectural contracts:
tests should fail if an implementation change breaks them.

## Startup

```mermaid
flowchart TD
    Start["Process starts"]
    Parse["Parse flags and subcommand"]
    Standalone{"doctor, keymap, version, or help?"}
    Print["Print result and exit"]
    LoadConfig["Load defaults and optional TOML"]
    Validate["Validate profile, theme, tab size, clipboard, bindings"]
    Screen["Create tcell screen"]
    Session["Create app session and command registry"]
    OpenArgs["Open each file or directory argument"]
    Run["Initialize terminal and run event loop"]
    Restore["Finalize screen and restore terminal"]

    Start --> Parse --> Standalone
    Standalone -->|yes| Print
    Standalone -->|no| LoadConfig --> Validate --> Screen --> Session --> OpenArgs --> Run --> Restore
```

An invalid user configuration fails before raw terminal mode begins. Once
`Run` initializes the screen, `Fini` is deferred so a normal return and a
reported panic restore terminal state.

## Input, command, edit, render

```mermaid
flowchart LR
    Raw["Terminal bytes"]
    Event["tcell event"]
    Normalize["Normalize key or paste"]
    Source{"Interaction source"}
    Resolve["Resolve keymap chord"]
    Menu["Bottom menu selection"]
    Palette["Palette selection"]
    Command["Central command ID"]
    Operation["Editor or session operation"]
    State["Buffer, cursor, document, or UI state"]
    Draw["Draw tabs, panes, overlays, menu, status"]

    Raw --> Event --> Normalize --> Source
    Source -->|keyboard| Resolve --> Command
    Source -->|bottom menu| Menu --> Command
    Source -->|palette| Palette --> Command
    Command --> Operation --> State --> Draw
```

The renderer never interprets editing shortcuts. The menu never edits text
directly. Both properties keep key profile changes independent of editor
logic.

## Anchored selection

```mermaid
stateDiagram-v2
    [*] --> Caret
    Caret --> ForwardSelection: Shift plus movement forward
    Caret --> BackwardSelection: Shift plus movement backward
    ForwardSelection --> Caret: head returns to anchor
    BackwardSelection --> Caret: head returns to anchor
    ForwardSelection --> BackwardSelection: head crosses anchor
    BackwardSelection --> ForwardSelection: head crosses anchor
    ForwardSelection --> Caret: non-selecting horizontal move collapses
    BackwardSelection --> Caret: non-selecting horizontal move collapses
```

The anchor remains fixed in all Shift transitions. Vertical transitions also
retain the desired terminal-cell column through short lines.

## Pending command-line file

`axam notes1.txt` binds an empty document to an absolute path without
creating a directory entry.

```mermaid
flowchart TD
    Arg["Relative argument notes1.txt"]
    Absolute["Resolve against process working directory"]
    Exists{"Path exists?"}
    Read["Read and decode existing file"]
    Parent{"Parent exists and is a directory?"}
    Error["Return a clear open error"]
    Pending["Open empty pending document; Exists=false"]
    Changed{"User changed the document?"}
    Close["Close document"]
    Save{"Explicit Save?"}
    Atomic["Create sibling temp, sync, chmod, rename"]
    Created["File now exists; mark editor clean"]
    Done["Done; no file created"]

    Arg --> Absolute --> Exists
    Exists -->|yes| Read
    Exists -->|no| Parent
    Parent -->|no| Error
    Parent -->|yes| Pending
    Pending --> Changed
    Changed -->|no| Close
    Close -->|"no filesystem operation"| Done
    Changed -->|yes| Save
    Save -->|yes| Atomic --> Created
    Save -->|"discard on close"| Done
```

Saving an unchanged pending document is still an explicit request and creates
an empty file. Merely opening and closing it does not.

## Atomic save and external-change guard

```mermaid
sequenceDiagram
    participant User
    participant App
    participant Guard as Disk change guard
    participant Backup
    participant Temp as Sibling temp file
    participant Target as Target file

    User->>App: Save
    App->>Guard: Compare mtime, size, and content hash
    alt target changed externally
        Guard-->>App: conflict
        App-->>User: overwrite confirmation
        User->>App: confirm or cancel
    else unchanged or new
        Guard-->>App: safe
    end
    opt backups enabled and target exists
        App->>Backup: copy old bytes to config backup directory
    end
    App->>Temp: create, write, sync, chmod
    App->>Target: rename temp over target
    App->>App: refresh hash and mark clean
```

On Unix, rename-over is atomic. The Windows fallback removes the destination
immediately before rename and therefore cannot make the same atomicity claim.
A save never truncates the target before a replacement is prepared.

## Clipboard paths

```mermaid
flowchart TD
    Copy["Copy or Cut command"]
    Internal["Update in-process clipboard"]
    Mode{"Configured clipboard mode"}
    NoSpawn["No child process"]
    OSC["tcell sends OSC 52 attempt"]
    Xclip["Run xclip"]
    Wayland["Run wl-copy"]
    Status["Report editor clipboard success; report external failure separately"]

    Copy --> Internal --> Mode
    Mode -->|internal| NoSpawn --> Status
    Mode -->|auto or osc52| OSC --> Status
    Mode -->|xclip| Xclip --> Status
    Mode -->|wl-copy| Wayland --> Status
```

OSC 52 transmission is not proof that a desktop clipboard changed. Over SSH,
the local terminal decides whether to accept it.

## Project discovery

```mermaid
flowchart LR
    Root["Opened directory or detected project root"]
    Walk["Bounded directory walk"]
    Ignore["Skip .git, generated directories, user patterns, directory symlinks"]
    Limit["Stop at 20,000 files"]
    List["Quick-open list"]
    Search["Project search"]
    HitLimit["Stop at 500 displayed hits"]

    Root --> Walk --> Ignore --> Limit --> List
    List --> Search --> HitLimit
```

Project search skips NUL-marked binary files and files above its per-file
limit. Long operations accept cancellation contexts even though the first
release currently invokes them synchronously from the session.
