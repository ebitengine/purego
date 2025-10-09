# Draft PR: Add embedded library loader utilities and example

## Summary
- add `OpenEmbeddedLibrary` helper that materializes embedded shared objects (Unix/Windows) and cleans them up safely
- provide cross-platform test coverage that embeds prebaked shared libraries and validates symbol invocation
- introduce `examples/embedlib` runnable example that embeds stock raylib binaries with go:embed, plus scripts/docs for refreshing them
- document platform support expectations via a matrix in the README (also pasted below for quick reference)

| Platform              | Status        | Notes                                                                                           |
|-----------------------|---------------|-------------------------------------------------------------------------------------------------|
| macOS (amd64/arm64)   | Supported     | Embeds the official raylib `libraylib.5.5.0.dylib`.                                             |
| Linux (amd64)         | Supported     | Includes the raylib `libraylib.so.5.5.0`; other architectures need prebuilt artifacts.          |
| Windows (amd64)       | Supported     | Embeds the raylib `raylib.dll`; Windows ignores `mode` flags.                                   |
| iOS (arm64)           | Not supported | iOS disallows loading unsigned binaries at runtime; use CGO and link at build time.             |
| Android               | Not yet tested| Would require packaging an Android-ready `.so` and executable temp storage.                     |

## Testing
- `go test ./...`
- `go run ./examples/embedlib`
