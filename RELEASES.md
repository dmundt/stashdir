# Creating Releases

This repository includes a GitHub Actions workflow to automatically build cross-platform binaries and create GitHub releases.

## Automatic Releases

When you push a new tag starting with `v` (e.g., `v1.0.1`, `v2.0.0`), the release workflow will automatically:

1. Build binaries for multiple platforms:
   - Linux (amd64, arm64)  
   - macOS (amd64, arm64)
   - Windows (amd64, arm64)
2. Create a GitHub release with the built binaries
3. Generate release notes automatically

## Manual Releases (for existing tags)

To create a release for an existing tag (like `v1.0.0`):

1. Go to the **Actions** tab in the GitHub repository
2. Select the **Release** workflow from the left sidebar
3. Click **Run workflow** button
4. Enter the tag name (e.g., `v1.0.0`) in the input field
5. Click **Run workflow**

The workflow will checkout the specified tag, build all binaries, and create the GitHub release.

## Binary Naming Convention

Built binaries follow this naming pattern:
- `stashdir-{os}-{arch}` for Unix-like systems
- `stashdir-{os}-{arch}.exe` for Windows

Examples:
- `stashdir-linux-amd64`
- `stashdir-darwin-arm64` 
- `stashdir-windows-amd64.exe`