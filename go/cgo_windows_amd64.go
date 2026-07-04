package oxigraph

// The vendored prebuilt library is preferred; a locally built
// target/x86_64-pc-windows-gnu/release library (cargo build -p
// oxigraph-ffi --release --target x86_64-pc-windows-gnu) is the
// development fallback. The library must be built for the GNU target —
// cgo links with MinGW, never MSVC. The system libraries mirror the
// staticlib's native-static-libs note. See lib/README.md.

/*
#cgo LDFLAGS: -L${SRCDIR}/lib/windows_amd64 -L${SRCDIR}/../target/x86_64-pc-windows-gnu/release -loxigraph_ffi -lrpcrt4 -lshlwapi -static-libgcc -static -lstdc++ -lkernel32 -lntdll -luserenv -lws2_32 -ldbghelp
*/
import "C"
