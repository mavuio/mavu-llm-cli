---
name: bit-ship
description: Ship modified bit components to bit.mavu.io (review diffs, tag, export)
---

# Bit Ship Workflow

version: 1.1.0

Export modified bit components to bit.mavu.io.

## Instructions

1. First, checkout the latest version of the workspace:
   ```bash
   bit checkout head --workspace-only
   ```

2. Check the status to get the list of modified components:
   ```bash
   bit status
   ```

3. Generate clean, shareable Bit diffs for review:
   - Create a clean diff folder (no ANSI color codes):
     ```bash
     mkdir -p /tmp/bit-diffs-clean
     ```
   - For each modified component, write `bit diff` output to a file and strip ANSI codes:
     ```bash
     NO_COLOR=1 bit diff <component_name> | perl -pe 's/\x1b\[[0-9;]*[A-Za-z]//g' > /tmp/bit-diffs-clean/<component_name>.diff
     ```
   - Open the folder in a new Cursor window:
     ```bash
     cursor /tmp/bit-diffs-clean
     ```
   - Alternative editors if needed:
     - `code /tmp/bit-diffs-clean`
     - Terminal review: `bit diff <component_name> | delta`

4. After the user reviews diffs, process each modified component:
   - Ask the user if they want to update this component
   - If yes, run: `bit tag <component_name> -m "<appropriate message based on the diff>"`
   - Continue with the next component

5. After all components are reviewed and tagged, export them:
   ```bash
   bit export
   ```

6. Verify everything is pushed:
   ```bash
   bit status
   ```

## Important Notes

- Always provide diffs from `bit diff` (not `git diff`) before asking for confirmation
- Prefer opening clean diff files in Cursor/VS Code; Fork does not reliably open standalone `.diff` files
- Generate an appropriate commit message based on the changes shown in the diff
- Wait for user confirmation before tagging each component
- Only export after all selected components have been tagged
