//! C ABI for embedding Oxigraph in other languages, per ADR 0001.
//!
//! The surface is deliberately coarse and string-based: whole operations
//! cross the boundary, structured data never does. Fallible functions
//! report failure through a `char **error_out` out-parameter instead of
//! thread-local state, because host runtimes with migrating green threads
//! (Go goroutines between cgo calls) cannot safely read back
//! thread-associated errors.
//!
//! Ownership conventions:
//! - Every `*mut OxigraphStore` returned by an `oxigraph_open*` function
//!   is owned by the caller and must be released exactly once with
//!   [`oxigraph_close`].
//! - Every `char *` the library writes into an out-parameter is owned by
//!   the caller and must be released with [`oxigraph_free_string`].
//! - The library never frees memory it did not allocate and never keeps
//!   pointers the caller passed in.
#![expect(unsafe_code)]

use oxigraph::io::{RdfFormat, RdfParser, RdfSerializer};
use oxigraph::model::Quad;
use oxigraph::sparql::results::{QueryResultsFormat, QueryResultsSerializer};
use oxigraph::sparql::{QueryResults, SparqlEvaluator, UpdateEvaluationError};
use oxigraph::store::{LoaderError, SerializerError, Store};
use std::ffi::{CStr, CString, c_char, c_int};
use std::fmt::Write;
use std::ptr;
use std::str::FromStr;

/// Query result kinds written to `kind_out` on success.
pub const OXIGRAPH_RESULT_SOLUTIONS: c_int = 1;
pub const OXIGRAPH_RESULT_BOOLEAN: c_int = 2;
pub const OXIGRAPH_RESULT_TRIPLES: c_int = 3;

/// Failure kinds written to `kind_out` when a call reports an error.
pub const OXIGRAPH_ERROR_SYNTAX: c_int = -1;
pub const OXIGRAPH_ERROR_EVALUATION: c_int = -2;
pub const OXIGRAPH_ERROR_STORAGE: c_int = -3;
pub const OXIGRAPH_ERROR_UNSUPPORTED_FORMAT: c_int = -4;

/// An opaque handle around an open [`Store`].
pub struct OxigraphStore {
    store: Store,
}

impl OxigraphStore {
    /// The wrapped store, for the query/update/load/dump slices.
    pub fn store(&self) -> &Store {
        &self.store
    }
}

/// Writes `message` into `error_out` as a caller-owned C string.
///
/// # Safety
///
/// `error_out` must be null or a valid pointer to writable memory.
unsafe fn set_error(error_out: *mut *mut c_char, message: &str) {
    if error_out.is_null() {
        return;
    }
    // A NUL inside an error message would truncate it; replace instead of
    // failing, so an error path can itself never fail.
    let message = CString::new(message.replace('\0', " ")).unwrap_or_default();
    // SAFETY: error_out is non-null and writable per the contract above.
    unsafe {
        *error_out = message.into_raw();
    }
}

/// Opens a read-write store backed by RocksDB at `path`, creating the
/// leaf directory if missing (parent directories are not created).
///
/// Returns null and writes a caller-owned message into `error_out` on
/// failure.
///
/// # Safety
///
/// `path` must be a valid NUL-terminated C string. `error_out` must be
/// null or a valid pointer to writable memory.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn oxigraph_open(
    path: *const c_char,
    error_out: *mut *mut c_char,
) -> *mut OxigraphStore {
    if path.is_null() {
        // SAFETY: error_out validity is the caller's contract.
        unsafe { set_error(error_out, "the store path must not be null") };
        return ptr::null_mut();
    }
    // SAFETY: path is a valid NUL-terminated C string per the contract.
    let path = unsafe { CStr::from_ptr(path) };
    let Ok(path) = path.to_str() else {
        // SAFETY: error_out validity is the caller's contract.
        unsafe { set_error(error_out, "the store path is not valid UTF-8") };
        return ptr::null_mut();
    };
    match Store::open(path) {
        Ok(store) => Box::into_raw(Box::new(OxigraphStore { store })),
        Err(e) => {
            // SAFETY: error_out validity is the caller's contract.
            unsafe { set_error(error_out, &e.to_string()) };
            ptr::null_mut()
        }
    }
}

/// Opens an in-memory store.
///
/// Returns null and writes a caller-owned message into `error_out` on
/// failure.
///
/// # Safety
///
/// `error_out` must be null or a valid pointer to writable memory.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn oxigraph_open_in_memory(
    error_out: *mut *mut c_char,
) -> *mut OxigraphStore {
    match Store::new() {
        Ok(store) => Box::into_raw(Box::new(OxigraphStore { store })),
        Err(e) => {
            // SAFETY: error_out validity is the caller's contract.
            unsafe { set_error(error_out, &e.to_string()) };
            ptr::null_mut()
        }
    }
}

/// Closes a store, releasing its resources (for an on-disk store, the
/// RocksDB directory lock). Null is tolerated; passing the same handle
/// twice is undefined behavior — the host binding must guard against it.
///
/// # Safety
///
/// `store` must be null or a pointer returned by an `oxigraph_open*`
/// function that has not been closed yet.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn oxigraph_close(store: *mut OxigraphStore) {
    if !store.is_null() {
        // SAFETY: store came from Box::into_raw in oxigraph_open* and is
        // dropped exactly once per the contract above. Dropping the last
        // owning Store releases the RocksDB lock.
        drop(unsafe { Box::from_raw(store) });
    }
}

/// Writes a result kind or failure kind into `kind_out`.
///
/// # Safety
///
/// `kind_out` must be null or a valid pointer to writable memory.
unsafe fn set_kind(kind_out: *mut c_int, kind: c_int) {
    if !kind_out.is_null() {
        // SAFETY: kind_out is non-null and writable per the contract above.
        unsafe {
            *kind_out = kind;
        }
    }
}

/// Evaluates a SPARQL query against the store and returns the fully
/// materialized result as a caller-owned string: SPARQL JSON results for
/// SELECT and ASK, N-Triples for CONSTRUCT and DESCRIBE. `kind_out`
/// receives the `OXIGRAPH_RESULT_*` kind on success.
///
/// Returns null on failure, writing an `OXIGRAPH_ERROR_*` kind into
/// `kind_out` and a caller-owned message into `error_out`.
///
/// # Safety
///
/// `store` must be a handle returned by an `oxigraph_open*` function that
/// has not been closed. `query` must be a valid NUL-terminated C string.
/// `kind_out` and `error_out` must each be null or a valid pointer to
/// writable memory.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn oxigraph_query(
    store: *const OxigraphStore,
    query: *const c_char,
    kind_out: *mut c_int,
    error_out: *mut *mut c_char,
) -> *mut c_char {
    // SAFETY: delegated to the caller's contract for every pointer.
    unsafe {
        let fail = |kind: c_int, message: &str| {
            set_kind(kind_out, kind);
            set_error(error_out, message);
            ptr::null_mut()
        };
        let Some(store) = store.as_ref() else {
            return fail(
                OXIGRAPH_ERROR_EVALUATION,
                "the store handle must not be null",
            );
        };
        if query.is_null() {
            return fail(OXIGRAPH_ERROR_SYNTAX, "the query must not be null");
        }
        let Ok(query) = CStr::from_ptr(query).to_str() else {
            return fail(OXIGRAPH_ERROR_SYNTAX, "the query is not valid UTF-8");
        };
        let parsed = match SparqlEvaluator::new().parse_query(query) {
            Ok(parsed) => parsed,
            Err(e) => return fail(OXIGRAPH_ERROR_SYNTAX, &e.to_string()),
        };
        let results = match parsed.on_store(&store.store).execute() {
            Ok(results) => results,
            Err(e) => return fail(OXIGRAPH_ERROR_EVALUATION, &e.to_string()),
        };
        match serialize_query_results(results) {
            Ok((payload, kind)) => {
                let Ok(payload) = CString::new(payload) else {
                    // The JSON and N-Triples serializers escape control
                    // characters, so an interior NUL cannot occur.
                    return fail(OXIGRAPH_ERROR_EVALUATION, "the result contains a NUL byte");
                };
                set_kind(kind_out, kind);
                payload.into_raw()
            }
            Err(message) => fail(OXIGRAPH_ERROR_EVALUATION, &message),
        }
    }
}

/// Serializes query results: SPARQL JSON for solutions and booleans,
/// N-Triples for graphs.
fn serialize_query_results(results: QueryResults<'_>) -> Result<(Vec<u8>, c_int), String> {
    match results {
        QueryResults::Solutions(solutions) => {
            let mut buffer = Vec::new();
            let mut serializer = QueryResultsSerializer::from_format(QueryResultsFormat::Json)
                .serialize_solutions_to_writer(&mut buffer, solutions.variables().to_vec())
                .map_err(|e| e.to_string())?;
            for solution in solutions {
                let solution = solution.map_err(|e| e.to_string())?;
                serializer.serialize(&solution).map_err(|e| e.to_string())?;
            }
            serializer.finish().map_err(|e| e.to_string())?;
            Ok((buffer, OXIGRAPH_RESULT_SOLUTIONS))
        }
        QueryResults::Boolean(value) => {
            let mut buffer = Vec::new();
            QueryResultsSerializer::from_format(QueryResultsFormat::Json)
                .serialize_boolean_to_writer(&mut buffer, value)
                .map_err(|e| e.to_string())?;
            Ok((buffer, OXIGRAPH_RESULT_BOOLEAN))
        }
        QueryResults::Graph(triples) => {
            let mut buffer = String::new();
            for triple in triples {
                let triple = triple.map_err(|e| e.to_string())?;
                writeln!(&mut buffer, "{triple} .").map_err(|e| e.to_string())?;
            }
            Ok((buffer.into_bytes(), OXIGRAPH_RESULT_TRIPLES))
        }
    }
}

/// Executes a SPARQL update against the store, applied atomically.
///
/// Returns 0 on success, or an `OXIGRAPH_ERROR_*` kind on failure with a
/// caller-owned message written into `error_out`.
///
/// # Safety
///
/// `store` must be a handle returned by an `oxigraph_open*` function that
/// has not been closed. `update` must be a valid NUL-terminated C string.
/// `error_out` must be null or a valid pointer to writable memory.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn oxigraph_update(
    store: *const OxigraphStore,
    update: *const c_char,
    error_out: *mut *mut c_char,
) -> c_int {
    // SAFETY: delegated to the caller's contract for every pointer.
    unsafe {
        let fail = |kind: c_int, message: &str| {
            set_error(error_out, message);
            kind
        };
        let Some(store) = store.as_ref() else {
            return fail(
                OXIGRAPH_ERROR_EVALUATION,
                "the store handle must not be null",
            );
        };
        if update.is_null() {
            return fail(OXIGRAPH_ERROR_SYNTAX, "the update must not be null");
        }
        let Ok(update) = CStr::from_ptr(update).to_str() else {
            return fail(OXIGRAPH_ERROR_SYNTAX, "the update is not valid UTF-8");
        };
        let parsed = match SparqlEvaluator::new().parse_update(update) {
            Ok(parsed) => parsed,
            Err(e) => return fail(OXIGRAPH_ERROR_SYNTAX, &e.to_string()),
        };
        match parsed.on_store(&store.store).execute() {
            Ok(()) => 0,
            Err(UpdateEvaluationError::Storage(e)) => fail(OXIGRAPH_ERROR_STORAGE, &e.to_string()),
            Err(e) => fail(OXIGRAPH_ERROR_EVALUATION, &e.to_string()),
        }
    }
}

/// Inserts the quad written as a single N-Quads statement line (the
/// trailing dot is optional). Inserting an already-present quad is a
/// no-op, per RDF set semantics.
///
/// Returns 0 on success, or an `OXIGRAPH_ERROR_*` kind on failure with a
/// caller-owned message written into `error_out`.
///
/// # Safety
///
/// `store` must be a handle returned by an `oxigraph_open*` function that
/// has not been closed. `quad` must be a valid NUL-terminated C string.
/// `error_out` must be null or a valid pointer to writable memory.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn oxigraph_add(
    store: *const OxigraphStore,
    quad: *const c_char,
    error_out: *mut *mut c_char,
) -> c_int {
    // SAFETY: delegated to the caller's contract for every pointer.
    unsafe {
        with_store_and_quad(store, quad, error_out, |store, quad| {
            store.insert(quad).map(|_| ())
        })
    }
}

/// Removes the quad written as a single N-Quads statement line (the
/// trailing dot is optional). Removing an absent quad is a no-op.
///
/// Returns 0 on success, or an `OXIGRAPH_ERROR_*` kind on failure with a
/// caller-owned message written into `error_out`.
///
/// # Safety
///
/// Same contract as [`oxigraph_add`].
#[unsafe(no_mangle)]
pub unsafe extern "C" fn oxigraph_remove(
    store: *const OxigraphStore,
    quad: *const c_char,
    error_out: *mut *mut c_char,
) -> c_int {
    // SAFETY: delegated to the caller's contract for every pointer.
    unsafe {
        with_store_and_quad(store, quad, error_out, |store, quad| {
            store.remove(&quad).map(|_| ())
        })
    }
}

/// Shared plumbing for [`oxigraph_add`] and [`oxigraph_remove`]: checks
/// the pointers, parses the N-Quads statement, and classifies errors.
///
/// # Safety
///
/// Same contract as [`oxigraph_add`].
unsafe fn with_store_and_quad(
    store: *const OxigraphStore,
    quad: *const c_char,
    error_out: *mut *mut c_char,
    operation: impl FnOnce(&Store, Quad) -> Result<(), oxigraph::store::StorageError>,
) -> c_int {
    // SAFETY: delegated to the caller's contract for every pointer.
    unsafe {
        let fail = |kind: c_int, message: &str| {
            set_error(error_out, message);
            kind
        };
        let Some(store) = store.as_ref() else {
            return fail(
                OXIGRAPH_ERROR_EVALUATION,
                "the store handle must not be null",
            );
        };
        if quad.is_null() {
            return fail(OXIGRAPH_ERROR_SYNTAX, "the quad must not be null");
        }
        let Ok(quad) = CStr::from_ptr(quad).to_str() else {
            return fail(OXIGRAPH_ERROR_SYNTAX, "the quad is not valid UTF-8");
        };
        let quad = match Quad::from_str(quad) {
            Ok(quad) => quad,
            Err(e) => return fail(OXIGRAPH_ERROR_SYNTAX, &e.to_string()),
        };
        match operation(&store.store, quad) {
            Ok(()) => 0,
            Err(e) => fail(OXIGRAPH_ERROR_STORAGE, &e.to_string()),
        }
    }
}

/// Resolves a format identifier (pyoxigraph's RdfFormat naming) to a
/// format.
fn rdf_format_from_name(name: &str) -> Option<RdfFormat> {
    match name {
        "Turtle" => Some(RdfFormat::Turtle),
        "N-Triples" => Some(RdfFormat::NTriples),
        "N-Quads" => Some(RdfFormat::NQuads),
        "TriG" => Some(RdfFormat::TriG),
        _ => None,
    }
}

/// Loads an RDF document into the store, atomically: either every quad
/// is inserted or none is. `format` is one of "Turtle", "N-Triples",
/// "N-Quads" or "TriG". `data` points to `len` bytes of document text
/// and may be null when `len` is 0.
///
/// Returns 0 on success, or an `OXIGRAPH_ERROR_*` kind on failure with a
/// caller-owned message written into `error_out`.
///
/// # Safety
///
/// `store` must be a handle returned by an `oxigraph_open*` function that
/// has not been closed. `format` must be a valid NUL-terminated C string.
/// `data` must be null (only when `len` is 0) or point to `len` readable
/// bytes. `error_out` must be null or a valid pointer to writable memory.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn oxigraph_load(
    store: *const OxigraphStore,
    format: *const c_char,
    data: *const c_char,
    len: usize,
    error_out: *mut *mut c_char,
) -> c_int {
    // SAFETY: delegated to the caller's contract for every pointer.
    unsafe {
        let fail = |kind: c_int, message: &str| {
            set_error(error_out, message);
            kind
        };
        let Some(store) = store.as_ref() else {
            return fail(
                OXIGRAPH_ERROR_EVALUATION,
                "the store handle must not be null",
            );
        };
        let Some(format) = parse_format(format) else {
            return fail(
                OXIGRAPH_ERROR_UNSUPPORTED_FORMAT,
                "unknown RDF format identifier",
            );
        };
        let data: &[u8] = if data.is_null() {
            if len != 0 {
                return fail(
                    OXIGRAPH_ERROR_EVALUATION,
                    "data must not be null when len is not zero",
                );
            }
            &[]
        } else if len == 0 {
            &[]
        } else {
            std::slice::from_raw_parts(data.cast(), len)
        };
        match store
            .store
            .load_from_slice(RdfParser::from_format(format), data)
        {
            Ok(()) => 0,
            Err(LoaderError::Parsing(e)) => fail(OXIGRAPH_ERROR_SYNTAX, &e.to_string()),
            Err(LoaderError::Storage(e)) => fail(OXIGRAPH_ERROR_STORAGE, &e.to_string()),
            Err(e) => fail(OXIGRAPH_ERROR_EVALUATION, &e.to_string()),
        }
    }
}

/// Serializes the whole store (default and named graphs) into a
/// caller-owned string. `format` must be a dataset format ("N-Quads" or
/// "TriG"); triples-only formats are rejected with
/// `OXIGRAPH_ERROR_UNSUPPORTED_FORMAT`, matching pyoxigraph's dump.
///
/// Returns null on failure, writing a caller-owned message into
/// `error_out`.
///
/// # Safety
///
/// Same contract as [`oxigraph_load`] for `store`, `format` and
/// `error_out`.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn oxigraph_dump(
    store: *const OxigraphStore,
    format: *const c_char,
    error_out: *mut *mut c_char,
) -> *mut c_char {
    // SAFETY: delegated to the caller's contract for every pointer.
    unsafe {
        let fail = |message: &str| {
            set_error(error_out, message);
            ptr::null_mut()
        };
        let Some(store) = store.as_ref() else {
            return fail("the store handle must not be null");
        };
        let Some(format) = parse_format(format) else {
            return fail("unknown RDF format identifier");
        };
        match store
            .store
            .dump_to_writer(RdfSerializer::from_format(format), Vec::new())
        {
            Ok(buffer) => match CString::new(buffer) {
                Ok(payload) => payload.into_raw(),
                Err(_) => fail("the dump contains a NUL byte"),
            },
            Err(SerializerError::DatasetFormatExpected(format)) => fail(&format!(
                "{format} stores triples only; dump the whole store with a dataset format such as N-Quads or TriG"
            )),
            Err(e) => fail(&e.to_string()),
        }
    }
}

/// Parses a format identifier C string.
///
/// # Safety
///
/// `format` must be null or a valid NUL-terminated C string.
unsafe fn parse_format(format: *const c_char) -> Option<RdfFormat> {
    if format.is_null() {
        return None;
    }
    // SAFETY: format is a valid NUL-terminated C string per the contract.
    let format = unsafe { CStr::from_ptr(format) };
    rdf_format_from_name(format.to_str().ok()?)
}

/// Frees a string the library wrote into an out-parameter. Null is
/// tolerated.
///
/// # Safety
///
/// `s` must be null or a string the library handed to the caller that has
/// not been freed yet.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn oxigraph_free_string(s: *mut c_char) {
    if !s.is_null() {
        // SAFETY: s came from CString::into_raw in set_error and is
        // freed exactly once per the contract above.
        drop(unsafe { CString::from_raw(s) });
    }
}

#[cfg(test)]
#[expect(clippy::panic_in_result_fn, clippy::unwrap_used)]
mod tests {
    use super::*;
    use std::fs;
    use std::path::PathBuf;
    use std::process;

    fn temp_dir(name: &str) -> PathBuf {
        let dir = std::env::temp_dir().join(format!("oxigraph-ffi-test-{}-{name}", process::id()));
        let _ = fs::remove_dir_all(&dir);
        dir
    }

    fn open(path: &std::path::Path) -> (*mut OxigraphStore, *mut c_char) {
        let c_path = CString::new(path.to_str().unwrap()).unwrap();
        let mut error: *mut c_char = ptr::null_mut();
        let store = unsafe { oxigraph_open(c_path.as_ptr(), &raw mut error) };
        (store, error)
    }

    fn error_text(error: *mut c_char) -> String {
        assert!(!error.is_null());
        let text = unsafe { CStr::from_ptr(error) }
            .to_str()
            .unwrap()
            .to_owned();
        unsafe { oxigraph_free_string(error) };
        text
    }

    #[test]
    fn open_creates_leaf_directory_and_close_releases_the_lock() {
        let dir = temp_dir("lifecycle");
        let (store, error) = open(&dir);
        assert!(!store.is_null(), "open failed: {}", error_text(error));
        assert!(dir.is_dir());
        unsafe { oxigraph_close(store) };

        let (reopened, error) = open(&dir);
        assert!(!reopened.is_null(), "reopen failed: {}", error_text(error));
        unsafe { oxigraph_close(reopened) };
        fs::remove_dir_all(&dir).unwrap();
    }

    #[test]
    fn open_locked_directory_fails() {
        let dir = temp_dir("locked");
        let (store, _) = open(&dir);
        assert!(!store.is_null());

        let (second, error) = open(&dir);
        assert!(second.is_null());
        let message = error_text(error);
        assert!(!message.is_empty());

        unsafe { oxigraph_close(store) };
        fs::remove_dir_all(&dir).unwrap();
    }

    #[test]
    fn open_missing_parent_fails() {
        let dir = temp_dir("missing-parent").join("leaf");
        let (store, error) = open(&dir);
        assert!(store.is_null());
        assert!(!error_text(error).is_empty());
    }

    #[test]
    fn open_null_path_fails() {
        let mut error: *mut c_char = ptr::null_mut();
        let store = unsafe { oxigraph_open(ptr::null(), &raw mut error) };
        assert!(store.is_null());
        assert_eq!(error_text(error), "the store path must not be null");
    }

    #[test]
    fn in_memory_store_opens_and_closes() {
        let mut error: *mut c_char = ptr::null_mut();
        let store = unsafe { oxigraph_open_in_memory(&raw mut error) };
        assert!(!store.is_null());
        unsafe { oxigraph_close(store) };
    }

    #[test]
    fn close_and_free_tolerate_null() {
        unsafe { oxigraph_close(ptr::null_mut()) };
        unsafe { oxigraph_free_string(ptr::null_mut()) };
    }

    fn run_query(
        store: *const OxigraphStore,
        query: &str,
    ) -> (Option<String>, c_int, Option<String>) {
        let c_query = CString::new(query).unwrap();
        let mut kind: c_int = 0;
        let mut error: *mut c_char = ptr::null_mut();
        let result =
            unsafe { oxigraph_query(store, c_query.as_ptr(), &raw mut kind, &raw mut error) };
        let payload = if result.is_null() {
            None
        } else {
            let text = unsafe { CStr::from_ptr(result) }
                .to_str()
                .unwrap()
                .to_owned();
            unsafe { oxigraph_free_string(result) };
            Some(text)
        };
        let message = if error.is_null() {
            None
        } else {
            Some(error_text(error))
        };
        (payload, kind, message)
    }

    #[test]
    fn query_returns_solutions_boolean_and_triples() {
        let mut error: *mut c_char = ptr::null_mut();
        let store = unsafe { oxigraph_open_in_memory(&raw mut error) };
        assert!(!store.is_null());

        let (payload, kind, _) = run_query(store, "SELECT (1 AS ?x) WHERE {}");
        assert_eq!(kind, OXIGRAPH_RESULT_SOLUTIONS);
        assert!(payload.unwrap().contains("\"x\""));

        let (payload, kind, _) = run_query(store, "ASK { FILTER(true) }");
        assert_eq!(kind, OXIGRAPH_RESULT_BOOLEAN);
        assert!(payload.unwrap().contains("true"));

        let (payload, kind, _) = run_query(
            store,
            "CONSTRUCT { <http://example.com/s> <http://example.com/p> \"o\" } WHERE {}",
        );
        assert_eq!(kind, OXIGRAPH_RESULT_TRIPLES);
        assert!(payload.unwrap().contains("<http://example.com/s>"));

        unsafe { oxigraph_close(store) };
    }

    fn run_statement(
        f: unsafe extern "C" fn(*const OxigraphStore, *const c_char, *mut *mut c_char) -> c_int,
        store: *const OxigraphStore,
        input: &str,
    ) -> (c_int, Option<String>) {
        let c_input = CString::new(input).unwrap();
        let mut error: *mut c_char = ptr::null_mut();
        let status = unsafe { f(store, c_input.as_ptr(), &raw mut error) };
        let message = if error.is_null() {
            None
        } else {
            Some(error_text(error))
        };
        (status, message)
    }

    #[test]
    fn add_update_and_remove_round_trip() {
        let mut error: *mut c_char = ptr::null_mut();
        let store = unsafe { oxigraph_open_in_memory(&raw mut error) };
        assert!(!store.is_null());

        let quad = "<http://example.com/s> <http://example.com/p> \"o\" .";
        let (status, message) = run_statement(oxigraph_add, store, quad);
        assert_eq!(status, 0, "add failed: {message:?}");

        let (payload, kind, _) = run_query(store, "ASK { <http://example.com/s> ?p ?o }");
        assert_eq!(kind, OXIGRAPH_RESULT_BOOLEAN);
        assert!(payload.unwrap().contains("true"));

        let (status, _) = run_statement(oxigraph_remove, store, quad);
        assert_eq!(status, 0);
        let (payload, _, _) = run_query(store, "ASK { ?s ?p ?o }");
        assert!(payload.unwrap().contains("false"));

        let (status, message) = run_statement(
            oxigraph_update,
            store,
            "INSERT DATA { <http://example.com/s2> <http://example.com/p> \"v\" }",
        );
        assert_eq!(status, 0, "update failed: {message:?}");
        let (payload, _, _) = run_query(store, "ASK { <http://example.com/s2> ?p ?o }");
        assert!(payload.unwrap().contains("true"));

        let (status, message) = run_statement(oxigraph_update, store, "INSRT DATA { }");
        assert_eq!(status, OXIGRAPH_ERROR_SYNTAX);
        assert!(!message.unwrap().is_empty());

        let (status, _) = run_statement(oxigraph_add, store, "not a quad");
        assert_eq!(status, OXIGRAPH_ERROR_SYNTAX);

        unsafe { oxigraph_close(store) };
    }

    #[test]
    fn load_and_dump_round_trip() {
        let mut error: *mut c_char = ptr::null_mut();
        let store = unsafe { oxigraph_open_in_memory(&raw mut error) };
        assert!(!store.is_null());

        let turtle = "<http://example.com/s> <http://example.com/p> \"o\" .";
        let c_format = CString::new("Turtle").unwrap();
        let mut load_error: *mut c_char = ptr::null_mut();
        let status = unsafe {
            oxigraph_load(
                store,
                c_format.as_ptr(),
                turtle.as_ptr().cast(),
                turtle.len(),
                &raw mut load_error,
            )
        };
        assert_eq!(status, 0);

        let c_nquads = CString::new("N-Quads").unwrap();
        let mut dump_error: *mut c_char = ptr::null_mut();
        let dump = unsafe { oxigraph_dump(store, c_nquads.as_ptr(), &raw mut dump_error) };
        assert!(!dump.is_null());
        let text = unsafe { CStr::from_ptr(dump) }.to_str().unwrap().to_owned();
        unsafe { oxigraph_free_string(dump) };
        assert!(text.contains("<http://example.com/s>"));

        // Triples-only dump is rejected.
        let mut turtle_error: *mut c_char = ptr::null_mut();
        let rejected = unsafe { oxigraph_dump(store, c_format.as_ptr(), &raw mut turtle_error) };
        assert!(rejected.is_null());
        assert!(error_text(turtle_error).contains("Turtle"));

        // A malformed document reports the line and loads nothing.
        let broken = "<http://example.com/s> <http://example.com/p> .";
        let mut broken_error: *mut c_char = ptr::null_mut();
        let status = unsafe {
            oxigraph_load(
                store,
                c_format.as_ptr(),
                broken.as_ptr().cast(),
                broken.len(),
                &raw mut broken_error,
            )
        };
        assert_eq!(status, OXIGRAPH_ERROR_SYNTAX);
        assert!(error_text(broken_error).contains("line"));

        unsafe { oxigraph_close(store) };
    }

    #[test]
    fn query_reports_syntax_and_evaluation_errors() {
        let mut error: *mut c_char = ptr::null_mut();
        let store = unsafe { oxigraph_open_in_memory(&raw mut error) };
        assert!(!store.is_null());

        let (payload, kind, message) = run_query(store, "SELCT ?x WHERE {}");
        assert!(payload.is_none());
        assert_eq!(kind, OXIGRAPH_ERROR_SYNTAX);
        assert!(!message.unwrap().is_empty());

        let (payload, kind, message) = run_query(
            store,
            "SELECT ?x WHERE { VALUES ?x { 1 } FILTER(<http://example.com/nofn>(?x)) }",
        );
        assert!(payload.is_none());
        assert_eq!(kind, OXIGRAPH_ERROR_EVALUATION);
        assert!(message.unwrap().contains("http://example.com/nofn"));

        unsafe { oxigraph_close(store) };
    }
}
