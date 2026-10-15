# ADR 0002: Put the discoverable command menu at the bottom

- Status: Accepted
- Date: 2026-10-15
- Release: 1.1.0

## Context

The command palette makes all commands searchable, but it is hidden until a
user knows F1, Alt+X, or a profile shortcut. Mainstream modeless editors expose
stable categories such as File, Edit, Selection, Find, View, and Goto.

The top row is already the document tab bar. Replacing it would reduce
multi-document clarity, and adding a second permanent top row would move the
text origin further from terminal convention. The status line already anchors
session information at the bottom.

Every visible control must execute real behavior. A decorative menu-shaped
hint strip would not meet that requirement.

## Decision

Reserve one row immediately above the status line for:

`File  Edit  Selection  Find  View  Goto  Help`

Keep document tabs on the top row. The menu:

- is always visible at supported sizes;
- opens upward so it never extends below the terminal;
- is accessible by mouse and F10;
- uses Left/Right for categories, Up/Down for items, Enter to execute, and
  Escape or F10 to close;
- shows the active profile’s shortest shortcut beside each item;
- stores only command IDs and invokes the same central handler table used by
  keymaps and the palette;
- reduces the editor viewport by one row and participates in all overlay and
  resize calculations.

The Help category includes shortcut, keymap, terminal compatibility, and
configuration commands. It is product help, not a claim of online help.

## Alternatives considered

### Put the menu above the tabs

Rejected because it consumes scarce vertical space at the visually dominant
edge and separates tabs from the terminal title.

### Replace the status line

Rejected because line/column, encoding, EOL, dirty, language, and profile
information is continuously useful.

### Show shortcut hints without interaction

Rejected because visible labels would imply controls that do not work.

### Use only the command palette

Rejected because first-time discoverability would still depend on knowing the
palette shortcut.

### Implement separate callbacks in the menu

Rejected because behavior would drift from keybindings and tests. Command IDs
are the architectural integration point.

## Consequences

Positive:

- Core commands are discoverable without memorizing a shortcut.
- Mouse users have a familiar command surface.
- Shortcut labels reflect the active keymap automatically.
- Tests can prove that every menu item is known and handled.

Costs and risks:

- The editing viewport loses one terminal row.
- Menu state and popup hit testing add a new application mode.
- Tiny terminals below the supported 80x24 size degrade sooner.
- Category curation must be updated when important commands are added.

## Verification

SimulationScreen tests assert:

- all menu command IDs are known and have handlers;
- F10 opens, arrows navigate, Enter executes, and Escape closes;
- mouse clicks open a category and execute an item;
- the row contains expected categories;
- an upward popup fits an 80x24 screen;
- resizing recomputes menu and status rows without overlap.

The generated terminal demonstration includes tabs at the top and both the
menu and status rows at the bottom.
