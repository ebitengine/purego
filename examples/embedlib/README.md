# Embedded Raylib Example

This example shows how to bundle prebuilt [raylib](https://www.raylib.com/) shared
libraries into a Go binary via `go:embed` and load them on demand with
`purego.OpenEmbeddedLibrary`.

The `libs` directory holds a subset of official raylib release artifacts
(macOS universal, Linux amd64, Windows amd64). To refresh them, run:

```sh
./scripts/fetch-raylib.sh
```

The script downloads the stock raylib archives from GitHub releases and copies
only the shared libraries we embed. Raylib is provided under the zlib license;
see the upstream LICENSE file for details.
