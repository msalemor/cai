pub fn build_source_file_list(target_folder: &str) -> Vec<String> {
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

// Helper function to apply file filters
pub fn apply_file_filters(
    files: &[String],
    skip_pattern: Option<&str>,
    include_pattern: Option<&str>,
) -> Vec<String> {
    let mut filtered = files.to_vec();

    // Apply skip pattern filter
    if let Some(skip) = skip_pattern {
        filtered.retain(|file| !matches_pattern(file, skip));
    }

    // Apply include pattern filter
    if let Some(include) = include_pattern {
        let patterns: Vec<&str> = include.split(',').map(|s| s.trim()).collect();
        filtered.retain(|file| {
            patterns
                .iter()
                .any(|pattern| matches_pattern(file, pattern))
        });
    }

    filtered
}

// Simple pattern matching for file extensions and basic patterns
fn matches_pattern(file: &str, pattern: &str) -> bool {
    if pattern.starts_with("*.") {
        let ext = &pattern[2..];
        file.ends_with(&format!(".{}", ext))
    } else if pattern.contains('*') {
        // Basic wildcard support - convert to regex-like matching
        let pattern_parts: Vec<&str> = pattern.split('*').collect();
        if pattern_parts.len() == 2 {
            file.starts_with(pattern_parts[0]) && file.ends_with(pattern_parts[1])
        } else {
            file.contains(pattern)
        }
    } else {
        file.contains(pattern)
    }
}
