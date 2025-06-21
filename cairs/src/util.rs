pub fn list_source_files(target_folder: &str) -> Vec<String> {
    let mut source_files = Vec::new();
    let exts = [
        "rs", "py", "js", "ts", "java", "cpp", "c", "cs", "go", "rb", "php", "swift", "kt", "scala",
    ];
    let walker = match walkdir::WalkDir::new(target_folder)
        .into_iter()
        .collect::<Vec<_>>()
    {
        files => files,
    };
    for entry in walker {
        if let Ok(entry) = entry {
            if entry.file_type().is_file() {
                if let Some(ext) = entry.path().extension().and_then(|e| e.to_str()) {
                    if exts.contains(&ext) {
                        if let Some(path_str) = entry.path().to_str() {
                            source_files.push(path_str.to_string());
                        }
                    }
                }
            }
        }
    }
    source_files
}
