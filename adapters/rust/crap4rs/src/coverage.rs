use std::collections::{HashMap, HashSet};
use std::fs;
use std::path::{Path, PathBuf};

use rustc_demangle::demangle;
use serde::Deserialize;
use serde_json::Value;

use crate::model::build_key;

#[derive(Debug, Deserialize)]
struct CoverageExport {
    data: Vec<CoverageData>,
}

#[derive(Debug, Deserialize)]
struct CoverageData {
    functions: Vec<CoverageFunction>,
}

#[derive(Debug, Deserialize)]
struct CoverageFunction {
    name: String,
    #[serde(default)]
    count: u64,
    #[serde(default)]
    filenames: Vec<String>,
    #[serde(default)]
    regions: Vec<Vec<Value>>,
}

#[derive(Debug)]
struct Region {
    start_line: u64,
    end_line: u64,
    count: u64,
    file_index: usize,
}

#[derive(Debug, Default)]
struct CoverageAccumulator {
    file: String,
    line: Option<usize>,
    total_lines: HashSet<u64>,
    covered_lines: HashSet<u64>,
    fallback_count: u64,
}

pub fn parse_llvm_cov_json(path: &Path, root: &Path) -> Result<HashMap<String, f64>, String> {
    let source = fs::read_to_string(path)
        .map_err(|error| format!("reading coverage JSON {}: {error}", path.display()))?;
    let export: CoverageExport = serde_json::from_str(&source)
        .map_err(|error| format!("parsing coverage JSON {}: {error}", path.display()))?;
    build_coverage_map(&export, root)
}

fn build_coverage_map(
    export: &CoverageExport,
    root: &Path,
) -> Result<HashMap<String, f64>, String> {
    let mut by_key: HashMap<String, CoverageAccumulator> = HashMap::new();

    for data in &export.data {
        for function in &data.functions {
            let Some(func_name) = normalize_function_name(&function.name) else {
                continue;
            };
            if function.filenames.is_empty() {
                continue;
            }

            let primary_file = normalize_path(&function.filenames[0], root);
            let mut primary_regions = Vec::new();
            let mut all_regions = Vec::new();
            for raw_region in &function.regions {
                let Some(region) = parse_region(raw_region) else {
                    continue;
                };
                if region.file_index == 0 {
                    primary_regions.push(region);
                } else {
                    all_regions.push(region);
                }
            }

            let selected_regions: Vec<Region> = if primary_regions.is_empty() {
                all_regions
            } else {
                primary_regions
            };

            let line = selected_regions
                .iter()
                .map(|region| region.start_line)
                .min()
                .unwrap_or(0) as usize;
            if line == 0 {
                continue;
            }

            let key = build_key(&primary_file, line, &func_name);
            let accumulator = by_key.entry(key).or_insert_with(|| CoverageAccumulator {
                file: primary_file.clone(),
                line: Some(line),
                total_lines: HashSet::new(),
                covered_lines: HashSet::new(),
                fallback_count: 0,
            });

            if accumulator.line.is_none() {
                accumulator.line = Some(line);
            }
            accumulator.fallback_count = accumulator.fallback_count.max(function.count);

            for region in selected_regions {
                add_region_lines(accumulator, &region);
            }
        }
    }

    let mut coverage = HashMap::new();
    for (key, accumulator) in by_key {
        if accumulator.file.is_empty() || accumulator.line.is_none() {
            continue;
        }
        let percent = if accumulator.total_lines.is_empty() {
            if accumulator.fallback_count > 0 {
                100.0
            } else {
                0.0
            }
        } else {
            (accumulator.covered_lines.len() as f64 / accumulator.total_lines.len() as f64) * 100.0
        };
        coverage.insert(key, percent);
    }

    Ok(coverage)
}

fn add_region_lines(accumulator: &mut CoverageAccumulator, region: &Region) {
    if region.end_line < region.start_line {
        return;
    }
    for line in region.start_line..=region.end_line {
        accumulator.total_lines.insert(line);
        if region.count > 0 {
            accumulator.covered_lines.insert(line);
        }
    }
}

fn parse_region(values: &[Value]) -> Option<Region> {
    if values.len() < 6 {
        return None;
    }
    let start_line = json_u64(&values[0])?;
    let end_line = json_u64(&values[2])?;
    let count = json_u64(&values[4])?;
    let file_index = json_u64(&values[5])? as usize;

    Some(Region {
        start_line,
        end_line,
        count,
        file_index,
    })
}

fn json_u64(value: &Value) -> Option<u64> {
    value
        .as_u64()
        .or_else(|| value.as_i64().and_then(|signed| u64::try_from(signed).ok()))
}

fn normalize_path(raw: &str, root: &Path) -> String {
    let root = fs::canonicalize(root).unwrap_or_else(|_| PathBuf::from(root));
    let candidate = PathBuf::from(raw);
    let normalized = if candidate.is_absolute() {
        candidate
            .strip_prefix(&root)
            .unwrap_or(&candidate)
            .to_path_buf()
    } else {
        candidate
    };
    normalized.to_string_lossy().replace('\\', "/")
}

fn normalize_function_name(raw: &str) -> Option<String> {
    let demangled = demangle(raw).to_string();
    let mut segments = demangled
        .split("::")
        .filter(|segment| !segment.is_empty())
        .collect::<Vec<_>>();
    if segments.is_empty() {
        return None;
    }
    if segments
        .last()
        .is_some_and(|segment| is_hash_segment(segment))
        && segments.len() > 1
    {
        segments.pop();
    }
    let last = segments.last()?.trim();
    if last.starts_with("{{closure") {
        return None;
    }
    let bare = last.split('<').next()?.trim();
    if bare.is_empty() {
        return None;
    }
    Some(bare.to_string())
}

fn is_hash_segment(segment: &str) -> bool {
    let bytes = segment.as_bytes();
    if bytes.len() != 17 || bytes[0] != b'h' {
        return false;
    }
    bytes[1..].iter().all(|byte| byte.is_ascii_hexdigit())
}

#[cfg(test)]
mod tests {
    use std::collections::HashMap;
    use std::path::PathBuf;

    use super::{build_coverage_map, CoverageData, CoverageExport, CoverageFunction};
    use crate::model::build_key;

    #[test]
    fn parses_function_coverage_percentages() {
        let export = CoverageExport {
            data: vec![CoverageData {
                functions: vec![
                    CoverageFunction {
                        name: "demo::alpha".to_string(),
                        count: 1,
                        filenames: vec!["src/lib.rs".to_string()],
                        regions: vec![vec![
                            3.into(),
                            1.into(),
                            3.into(),
                            10.into(),
                            1.into(),
                            0.into(),
                        ]],
                    },
                    CoverageFunction {
                        name: "demo::beta".to_string(),
                        count: 0,
                        filenames: vec!["src/lib.rs".to_string()],
                        regions: vec![vec![
                            10.into(),
                            1.into(),
                            10.into(),
                            10.into(),
                            0.into(),
                            0.into(),
                        ]],
                    },
                ],
            }],
        };

        let coverage = build_coverage_map(&export, &PathBuf::from(".")).expect("build coverage");
        let mut expected = HashMap::new();
        expected.insert(build_key("src/lib.rs", 3, "alpha"), 100.0);
        expected.insert(build_key("src/lib.rs", 10, "beta"), 0.0);

        assert_eq!(coverage, expected);
    }
}
