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

use oxigraph::store::Store;
use std::ffi::{CStr, CString, c_char};
use std::ptr;

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
pub unsafe extern "C" fn oxigraph_open_in_memory(error_out: *mut *mut c_char) -> *mut OxigraphStore {
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
        let text = unsafe { CStr::from_ptr(error) }.to_str().unwrap().to_owned();
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
}
