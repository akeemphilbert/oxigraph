package oxigraph

// The vendored prebuilt library is preferred; a locally built
// target/release library (cargo build -p oxigraph-ffi --release) is the
// development fallback. See lib/README.md.

/*
#cgo LDFLAGS: -L${SRCDIR}/lib/darwin_arm64 -L${SRCDIR}/../target/release -loxigraph_ffi -lc++
*/
import "C"
