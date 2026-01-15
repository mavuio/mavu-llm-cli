---
name: bit-ship
description: Ship modified bit components to bit.mavu.io (review diffs, tag, export)
---

# Bit Ship Workflow

version: 1.0.1

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

3. For each modified component:
   - Show the diff: `bit diff <component_name>`
   - Ask the user if they want to update this component
   - If yes, run: `bit tag <component_name> -m "<appropriate message based on the diff>"`
   - Continue with the next component

4. After all components are reviewed and tagged, export them:
   ```bash
   bit export
   ```

5. Verify everything is pushed:
   ```bash
   bit status
   ```

## Important Notes

- Always show the diff before asking for confirmation
- Generate an appropriate commit message based on the changes shown in the diff
- Wait for user confirmation before tagging each component
- Only export after all selected components have been tagged
