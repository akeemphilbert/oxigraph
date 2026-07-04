/* C ABI for embedding Oxigraph, per ADR 0001.
 *
 * Ownership conventions:
 * - Every OxigraphStore* returned by an oxigraph_open* function is owned
 *   by the caller and must be released exactly once with oxigraph_close.
 * - Every char* the library writes into an out-parameter is owned by the
 *   caller and must be released with oxigraph_free_string.
 *
 * Fallible functions return NULL on failure and, when error_out is not
 * NULL, write a caller-owned message into *error_out.
 */
#ifndef OXIGRAPH_FFI_H
#define OXIGRAPH_FFI_H

#ifdef __cplusplus
extern "C" {
#endif

typedef struct OxigraphStore OxigraphStore;

/* Opens a read-write store backed by RocksDB at path, creating the leaf
 * directory if missing (parent directories are not created). */
OxigraphStore *oxigraph_open(const char *path, char **error_out);

/* Opens an in-memory store. */
OxigraphStore *oxigraph_open_in_memory(char **error_out);

/* Closes a store, releasing its resources (for an on-disk store, the
 * RocksDB directory lock). NULL is tolerated; closing the same handle
 * twice is undefined behavior — the host binding must guard against it. */
void oxigraph_close(OxigraphStore *store);

/* Frees a string the library wrote into an out-parameter. NULL is
 * tolerated. */
void oxigraph_free_string(char *s);

#ifdef __cplusplus
}
#endif

#endif /* OXIGRAPH_FFI_H */
