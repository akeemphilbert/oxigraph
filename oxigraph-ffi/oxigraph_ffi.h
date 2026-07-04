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

/* Query result kinds written to kind_out on success. */
#define OXIGRAPH_RESULT_SOLUTIONS 1
#define OXIGRAPH_RESULT_BOOLEAN 2
#define OXIGRAPH_RESULT_TRIPLES 3

/* Failure kinds written to kind_out when a call reports an error. */
#define OXIGRAPH_ERROR_SYNTAX (-1)
#define OXIGRAPH_ERROR_EVALUATION (-2)
#define OXIGRAPH_ERROR_STORAGE (-3)

/* Opens a read-write store backed by RocksDB at path, creating the leaf
 * directory if missing (parent directories are not created). */
OxigraphStore *oxigraph_open(const char *path, char **error_out);

/* Opens an in-memory store. */
OxigraphStore *oxigraph_open_in_memory(char **error_out);

/* Closes a store, releasing its resources (for an on-disk store, the
 * RocksDB directory lock). NULL is tolerated; closing the same handle
 * twice is undefined behavior — the host binding must guard against it. */
void oxigraph_close(OxigraphStore *store);

/* Evaluates a SPARQL query and returns the fully materialized result as
 * a caller-owned string: SPARQL JSON results for SELECT and ASK,
 * N-Triples for CONSTRUCT and DESCRIBE. *kind_out receives the
 * OXIGRAPH_RESULT_* kind. Returns NULL on failure, writing an
 * OXIGRAPH_ERROR_* kind into *kind_out and a caller-owned message into
 * *error_out. */
char *oxigraph_query(const OxigraphStore *store, const char *query,
                     int *kind_out, char **error_out);

/* Executes a SPARQL update, applied atomically. Returns 0 on success or
 * an OXIGRAPH_ERROR_* kind on failure, writing a caller-owned message
 * into *error_out. */
int oxigraph_update(const OxigraphStore *store, const char *update,
                    char **error_out);

/* Inserts the quad written as a single N-Quads statement line (trailing
 * dot optional). Inserting an already-present quad is a no-op. Same
 * return convention as oxigraph_update. */
int oxigraph_add(const OxigraphStore *store, const char *quad,
                 char **error_out);

/* Removes the quad written as a single N-Quads statement line (trailing
 * dot optional). Removing an absent quad is a no-op. Same return
 * convention as oxigraph_update. */
int oxigraph_remove(const OxigraphStore *store, const char *quad,
                    char **error_out);

/* Frees a string the library wrote into an out-parameter. NULL is
 * tolerated. */
void oxigraph_free_string(char *s);

#ifdef __cplusplus
}
#endif

#endif /* OXIGRAPH_FFI_H */
