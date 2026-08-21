use std::cmp::Ordering;
use std::io::Write;
use std::path::{Path, PathBuf};
use std::process::Command;

use serde::Serialize;

use crate::complexity::scan_dir;
use crate::coverage::parse_llvm_cov_json;
use crate::evaluate::evaluate;
use crate::model::Report;
use crate::thresholds::read_crap_ceiling;

#[derive(Clone, Copy, Eq, PartialEq)]
enum OutputFormat {
    Text,
    Json,
}

struct Options {
    dir: PathBuf,
    coverage_json: PathBuf,
    ceiling_file: PathBuf,
    ceiling_override: Option<f64>,
    changed: Option<String>,
    format: OutputFormat,
}

#[derive(Serialize)]
struct JsonFunctionReport {
    file: String,
    line: usize,
    #[serde(rename = "func")]
    func_name: String,
    complexity: u64,
    coverage: f64,
    crap: f64,
    pass: bool,
}

#[derive(Serialize)]
struct JsonSummary {
    total: usize,
    failing: usize,
    max_crap: f64,
}

#[derive(Serialize)]
struct JsonReport {
    ceiling: f64,
    functions: Vec<JsonFunctionReport>,
    summary: JsonSummary,
}

pub fn run(args: &[String], stdout: &mut dyn Write, stderr: &mut dyn Write) -> i32 {
    let options = match parse_options(args) {
        Ok(options) => options,
        Err(message) => {
            let _ = writeln!(stderr, "crap4rs: {message}");
            return 2;
        }
    };

    match run_with_options(options, stdout, stderr) {
        Ok(code) => code,
        Err(message) => {
            let _ = writeln!(stderr, "crap4rs: {message}");
            2
        }
    }
}

fn run_with_options(
    options: Options,
    stdout: &mut dyn Write,
    stderr: &mut dyn Write,
) -> Result<i32, String> {
    let root = std::fs::canonicalize(&options.dir)
        .map_err(|error| format!("resolving scan root {}: {error}", options.dir.display()))?;
    let coverage_json_path = resolve_path(&options.coverage_json, &root);
    let ceiling_file_path = resolve_path(&options.ceiling_file, &root);
    let mut functions = scan_dir(&root)?;
    functions.sort_by(sort_complexity_records);
    if let Some(reference) = &options.changed {
        let changed_files = git_changed_files(&root, reference)?;
        if changed_files.is_empty() {
            let _ = writeln!(
                stderr,
                "crap4rs: no files changed since {reference:?}; nothing was scored"
            );
        }
        functions.retain(|function| {
            changed_files
                .iter()
                .any(|candidate| files_match(&function.file, candidate))
        });
    }

    let coverage = parse_llvm_cov_json(&coverage_json_path, &root)?;
    let ceiling = if let Some(override_value) = options.ceiling_override {
        override_value
    } else {
        read_crap_ceiling(&ceiling_file_path)?
    };

    let report = evaluate(&functions, &coverage, ceiling);
    print_report(stdout, &report, options.format)?;

    if report.summary.failing > 0 {
        Ok(1)
    } else {
        Ok(0)
    }
}

fn sort_complexity_records(
    left: &crate::model::FunctionComplexity,
    right: &crate::model::FunctionComplexity,
) -> Ordering {
    left.file
        .cmp(&right.file)
        .then_with(|| left.line.cmp(&right.line))
        .then_with(|| left.func_name.cmp(&right.func_name))
}

fn print_report(
    stdout: &mut dyn Write,
    report: &Report,
    format: OutputFormat,
) -> Result<(), String> {
    match format {
        OutputFormat::Json => print_json(stdout, report),
        OutputFormat::Text => {
            print_text(stdout, report);
            Ok(())
        }
    }
}

fn print_json(stdout: &mut dyn Write, report: &Report) -> Result<(), String> {
    let encoded = JsonReport {
        ceiling: report.ceiling,
        functions: report
            .functions
            .iter()
            .map(|function| JsonFunctionReport {
                file: function.file.clone(),
                line: function.line,
                func_name: function.func_name.clone(),
                complexity: function.complexity,
                coverage: function.coverage,
                crap: function.crap,
                pass: function.pass,
            })
            .collect(),
        summary: JsonSummary {
            total: report.summary.total,
            failing: report.summary.failing,
            max_crap: report.summary.max_crap,
        },
    };

    serde_json::to_writer_pretty(&mut *stdout, &encoded)
        .map_err(|error| format!("encoding JSON report: {error}"))?;
    writeln!(stdout).map_err(|error| format!("writing JSON report: {error}"))
}

fn print_text(stdout: &mut dyn Write, report: &Report) {
    for function in &report.functions {
        let status = if function.pass { "PASS" } else { "FAIL" };
        let note = if function.matched {
            ""
        } else {
            " (no coverage data)"
        };
        let _ = writeln!(
            stdout,
            "{}:{}:{}\tcomplexity={}\tcoverage={:.1}%\tcrap={:.2}\t{}{}",
            function.file,
            function.line,
            function.func_name,
            function.complexity,
            function.coverage,
            function.crap,
            status,
            note
        );
    }
    let _ = writeln!(
        stdout,
        "\nceiling={:.2} total={} failing={} max_crap={:.2}",
        report.ceiling, report.summary.total, report.summary.failing, report.summary.max_crap
    );
}

fn parse_options(args: &[String]) -> Result<Options, String> {
    let mut dir = PathBuf::from(".");
    // Matches what the documented `cargo llvm-cov --json --output-path` writes.
    // The old default of "coverage.json" was never the path any doc produced.
    let mut coverage_json = PathBuf::from("target/llvm-cov.json");
    let mut ceiling_file = PathBuf::from(".gauntlet/thresholds.yml");
    let mut ceiling_override = None;
    let mut changed = None;
    let mut format = OutputFormat::Text;

    let mut index = 0;
    while index < args.len() {
        match args[index].as_str() {
            "--dir" => {
                index += 1;
                dir = parse_path_arg(args, index, "--dir")?;
            }
            // --coverage-json and --ceiling-file are the pre-v0.2.0 names.
            "--coverage" | "--coverage-json" => {
                index += 1;
                coverage_json = parse_path_arg(args, index, "--coverage")?;
            }
            "--thresholds" | "--ceiling-file" => {
                index += 1;
                ceiling_file = parse_path_arg(args, index, "--thresholds")?;
            }
            "--changed" => {
                index += 1;
                changed = Some(parse_string_arg(args, index, "--changed")?);
            }
            "--ceiling" => {
                index += 1;
                ceiling_override = Some(parse_f64_arg(args, index, "--ceiling")?);
            }
            "--format" => {
                index += 1;
                format = match parse_string_arg(args, index, "--format")?.as_str() {
                    "json" => OutputFormat::Json,
                    "text" => OutputFormat::Text,
                    value => {
                        return Err(format!(
                            "invalid --format {value:?}, expected \"text\" or \"json\""
                        ))
                    }
                };
            }
            "-h" | "--help" => {
                return Err(help_text().to_string());
            }
            unknown => return Err(format!("unknown flag {unknown:?}\n\n{}", help_text())),
        }
        index += 1;
    }

    Ok(Options {
        dir,
        coverage_json,
        ceiling_file,
        ceiling_override,
        changed,
        format,
    })
}

fn resolve_path(path: &Path, root: &Path) -> PathBuf {
    if path.is_absolute() {
        return path.to_path_buf();
    }
    root.join(path)
}

fn parse_path_arg(args: &[String], index: usize, flag: &str) -> Result<PathBuf, String> {
    parse_string_arg(args, index, flag).map(PathBuf::from)
}

fn parse_f64_arg(args: &[String], index: usize, flag: &str) -> Result<f64, String> {
    let value = parse_string_arg(args, index, flag)?;
    value
        .parse::<f64>()
        .map_err(|_| format!("invalid {flag} value {value:?}"))
}

fn parse_string_arg(args: &[String], index: usize, flag: &str) -> Result<String, String> {
    args.get(index)
        .cloned()
        .ok_or_else(|| format!("missing value for {flag}"))
}

fn help_text() -> &'static str {
    "usage: crap4rs [--dir <dir>] [--coverage <path>] [--thresholds <path>] [--ceiling <n>] [--changed <ref>] [--format text|json]"
}

/// Files touched since the merge base of `reference` and HEAD, plus untracked
/// ones.
fn git_changed_files(root: &Path, reference: &str) -> Result<Vec<String>, String> {
    let base = git_output(root, &["merge-base", reference, "HEAD"])?;
    let diff = git_output(root, &["diff", "--name-only", base.trim()])?;
    let untracked = git_output(root, &["ls-files", "--others", "--exclude-standard"])?;
    Ok(diff
        .lines()
        .chain(untracked.lines())
        .map(str::trim)
        .filter(|line| !line.is_empty())
        .map(str::to_string)
        .collect())
}

fn git_output(root: &Path, args: &[&str]) -> Result<String, String> {
    let output = Command::new("git")
        .args(args)
        .current_dir(root)
        .output()
        .map_err(|error| format!("running git {}: {error}", args.join(" ")))?;
    if !output.status.success() {
        return Err(format!(
            "git {}: {}",
            args.join(" "),
            String::from_utf8_lossy(&output.stderr).trim()
        ));
    }
    String::from_utf8(output.stdout)
        .map_err(|error| format!("git {} produced invalid UTF-8: {error}", args.join(" ")))
}

/// Compare two paths allowing a path-segment suffix match either way, because
/// `git diff --name-only` reports paths from the repo root while scanned files
/// are relative to --dir, which may sit below it.
fn files_match(left: &str, right: &str) -> bool {
    let left = left.replace('\\', "/");
    let right = right.replace('\\', "/");
    left == right || left.ends_with(&format!("/{right}")) || right.ends_with(&format!("/{left}"))
}

#[cfg(test)]
mod tests {
    use std::fs;
    use std::io::Cursor;
    use std::path::PathBuf;
    use std::sync::atomic::{AtomicU64, Ordering};
    use std::time::{SystemTime, UNIX_EPOCH};

    use serde_json::Value;

    use super::run;

    #[test]
    fn emits_json_report_shape() {
        let fixture = write_fixture_project();
        let coverage_path = fixture.join("coverage.json");
        let thresholds_path = fixture.join(".gauntlet/thresholds.yml");

        let mut stdout = Cursor::new(Vec::new());
        let mut stderr = Cursor::new(Vec::new());
        let args = vec![
            "--dir".to_string(),
            fixture.to_string_lossy().to_string(),
            "--coverage-json".to_string(),
            coverage_path.to_string_lossy().to_string(),
            "--ceiling-file".to_string(),
            thresholds_path.to_string_lossy().to_string(),
            "--format".to_string(),
            "json".to_string(),
        ];

        let code = run(&args, &mut stdout, &mut stderr);
        assert_eq!(code, 0, "stderr={}", bytes_to_string(stderr.get_ref()));

        let report: Value = serde_json::from_slice(stdout.get_ref()).expect("valid json");
        assert!(report.get("ceiling").is_some());
        let functions = report
            .get("functions")
            .and_then(Value::as_array)
            .expect("functions array");
        assert_eq!(functions.len(), 2);
        let coverage = functions[0]
            .get("coverage")
            .and_then(Value::as_f64)
            .expect("numeric coverage");
        assert_eq!(coverage, 100.0);
        assert!(report.get("summary").is_some());
    }

    #[test]
    fn exits_one_when_failing_above_ceiling() {
        let fixture = write_fixture_project();
        let coverage_path = fixture.join("coverage-none.json");
        let thresholds_path = fixture.join(".gauntlet/thresholds.yml");

        let mut stdout = Cursor::new(Vec::new());
        let mut stderr = Cursor::new(Vec::new());
        let args = vec![
            "--dir".to_string(),
            fixture.to_string_lossy().to_string(),
            "--coverage-json".to_string(),
            coverage_path.to_string_lossy().to_string(),
            "--ceiling-file".to_string(),
            thresholds_path.to_string_lossy().to_string(),
        ];

        let code = run(&args, &mut stdout, &mut stderr);
        assert_eq!(code, 1, "stderr={}", bytes_to_string(stderr.get_ref()));
    }

    #[test]
    fn ignores_vendored_gauntlet_checkout_in_consumer_repo() {
        let fixture = write_fixture_project();
        let vendored_src = fixture.join(".gauntlet-tools/gauntlet/adapters/rust/crap4rs/src");
        fs::create_dir_all(&vendored_src).expect("create vendored gauntlet source directory");
        fs::write(
            vendored_src.join("lib.rs"),
            "pub fn vendored_failure(v: i32) -> i32 {\n    if v > 0 { 1 } else if v < 0 { -1 } else { 0 }\n}\n",
        )
        .expect("write vendored gauntlet source");

        let mut stdout = Cursor::new(Vec::new());
        let mut stderr = Cursor::new(Vec::new());
        let args = vec![
            "--dir".to_string(),
            fixture.to_string_lossy().to_string(),
            "--coverage-json".to_string(),
            fixture
                .join("coverage-none.json")
                .to_string_lossy()
                .to_string(),
            "--ceiling".to_string(),
            "2".to_string(),
            "--format".to_string(),
            "json".to_string(),
        ];

        let code = run(&args, &mut stdout, &mut stderr);
        assert_eq!(code, 1, "stderr={}", bytes_to_string(stderr.get_ref()));

        let report: Value = serde_json::from_slice(stdout.get_ref()).expect("valid json");
        let functions = report["functions"].as_array().expect("functions array");
        assert_eq!(functions.len(), 2);
        assert_eq!(functions[0]["file"], "src/lib.rs");
        assert_eq!(functions[0]["func"], "hot");
        assert_eq!(functions[0]["pass"], false);
        assert_eq!(functions[1]["file"], "src/lib.rs");
        assert_eq!(functions[1]["func"], "cold");
        assert_eq!(functions[1]["pass"], true);
    }

    #[test]
    fn accepts_canonical_coverage_and_thresholds_flags() {
        let fixture = write_fixture_project();
        let mut stdout = Cursor::new(Vec::new());
        let mut stderr = Cursor::new(Vec::new());
        let args = vec![
            "--dir".to_string(),
            fixture.to_string_lossy().to_string(),
            "--coverage".to_string(),
            fixture.join("coverage.json").to_string_lossy().to_string(),
            "--thresholds".to_string(),
            fixture
                .join(".gauntlet/thresholds.yml")
                .to_string_lossy()
                .to_string(),
            "--format".to_string(),
            "json".to_string(),
        ];

        let code = run(&args, &mut stdout, &mut stderr);
        assert_eq!(code, 0, "stderr={}", bytes_to_string(stderr.get_ref()));
        let report: Value = serde_json::from_slice(stdout.get_ref()).expect("valid json");
        assert_eq!(report["functions"].as_array().expect("array").len(), 2);
    }

    #[test]
    fn ceiling_flag_overrides_the_thresholds_file() {
        let fixture = write_fixture_project();
        let mut stdout = Cursor::new(Vec::new());
        let mut stderr = Cursor::new(Vec::new());
        let args = vec![
            "--dir".to_string(),
            fixture.to_string_lossy().to_string(),
            "--coverage".to_string(),
            fixture
                .join("coverage-none.json")
                .to_string_lossy()
                .to_string(),
            "--thresholds".to_string(),
            fixture
                .join(".gauntlet/thresholds.yml")
                .to_string_lossy()
                .to_string(),
            "--ceiling".to_string(),
            "30".to_string(),
        ];

        let code = run(&args, &mut stdout, &mut stderr);
        assert_eq!(code, 0, "stderr={}", bytes_to_string(stderr.get_ref()));
    }

    #[test]
    fn missing_ceiling_source_is_a_usage_error() {
        let fixture = write_fixture_project();
        let mut stdout = Cursor::new(Vec::new());
        let mut stderr = Cursor::new(Vec::new());
        let args = vec![
            "--dir".to_string(),
            fixture.to_string_lossy().to_string(),
            "--coverage".to_string(),
            fixture.join("coverage.json").to_string_lossy().to_string(),
            "--thresholds".to_string(),
            fixture.join("absent.yml").to_string_lossy().to_string(),
        ];

        assert_eq!(run(&args, &mut stdout, &mut stderr), 2);
    }

    #[test]
    fn missing_coverage_file_is_a_usage_error() {
        let fixture = write_fixture_project();
        let mut stdout = Cursor::new(Vec::new());
        let mut stderr = Cursor::new(Vec::new());
        let args = vec![
            "--dir".to_string(),
            fixture.to_string_lossy().to_string(),
            "--coverage".to_string(),
            fixture.join("absent.json").to_string_lossy().to_string(),
            "--ceiling".to_string(),
            "30".to_string(),
        ];

        assert_eq!(run(&args, &mut stdout, &mut stderr), 2);
    }

    #[test]
    fn changed_scopes_the_report_to_touched_files() {
        let fixture = write_git_fixture_project(false);
        let mut stdout = Cursor::new(Vec::new());
        let mut stderr = Cursor::new(Vec::new());
        let args = vec![
            "--dir".to_string(),
            fixture.to_string_lossy().to_string(),
            "--coverage".to_string(),
            fixture
                .join("coverage-none.json")
                .to_string_lossy()
                .to_string(),
            "--ceiling".to_string(),
            "30".to_string(),
            "--changed".to_string(),
            "HEAD".to_string(),
            "--format".to_string(),
            "json".to_string(),
        ];

        let code = run(&args, &mut stdout, &mut stderr);
        assert_eq!(code, 0, "stderr={}", bytes_to_string(stderr.get_ref()));
        let report: Value = serde_json::from_slice(stdout.get_ref()).expect("valid json");
        let files: Vec<String> = report["functions"]
            .as_array()
            .expect("array")
            .iter()
            .map(|function| function["file"].as_str().expect("file").to_string())
            .collect();
        assert_eq!(files, vec!["src/touched.rs".to_string()]);
    }

    #[test]
    fn changed_with_an_empty_diff_warns_instead_of_passing_silently() {
        let fixture = write_git_fixture_project(true);
        let mut stdout = Cursor::new(Vec::new());
        let mut stderr = Cursor::new(Vec::new());
        let args = vec![
            "--dir".to_string(),
            fixture.to_string_lossy().to_string(),
            "--coverage".to_string(),
            fixture
                .join("coverage-none.json")
                .to_string_lossy()
                .to_string(),
            "--ceiling".to_string(),
            "1".to_string(),
            "--changed".to_string(),
            "HEAD".to_string(),
            "--format".to_string(),
            "json".to_string(),
        ];

        let code = run(&args, &mut stdout, &mut stderr);
        assert_eq!(code, 0, "stderr={}", bytes_to_string(stderr.get_ref()));
        assert!(
            bytes_to_string(stderr.get_ref()).contains("no files changed"),
            "stderr={}",
            bytes_to_string(stderr.get_ref())
        );
        let report: Value = serde_json::from_slice(stdout.get_ref()).expect("valid json");
        assert_eq!(report["summary"]["total"], 0);
    }

    #[test]
    fn changed_with_an_unknown_ref_is_a_usage_error() {
        let fixture = write_git_fixture_project(false);
        let mut stdout = Cursor::new(Vec::new());
        let mut stderr = Cursor::new(Vec::new());
        let args = vec![
            "--dir".to_string(),
            fixture.to_string_lossy().to_string(),
            "--coverage".to_string(),
            fixture
                .join("coverage-none.json")
                .to_string_lossy()
                .to_string(),
            "--ceiling".to_string(),
            "30".to_string(),
            "--changed".to_string(),
            "no-such-ref".to_string(),
        ];

        assert_eq!(run(&args, &mut stdout, &mut stderr), 2);
        assert!(bytes_to_string(stderr.get_ref()).contains("merge-base"));
    }

    #[test]
    fn files_match_allows_a_suffix_in_either_direction() {
        assert!(super::files_match("src/lib.rs", "src/lib.rs"));
        assert!(super::files_match("lib.rs", "src/lib.rs"));
        assert!(super::files_match("src/lib.rs", "lib.rs"));
        assert!(!super::files_match("mylib.rs", "lib.rs"));
        assert!(!super::files_match("src/a.rs", "src/b.rs"));
    }

    fn write_git_fixture_project(commit_everything: bool) -> PathBuf {
        let root = write_fixture_project();
        git(&root, &["init", "-q"]);
        git(&root, &["add", "."]);
        commit(&root, "initial");
        fs::write(
            root.join("src/touched.rs"),
            "pub fn touched(v: i32) -> i32 {\n    if v > 0 { 1 } else { 0 }\n}\n",
        )
        .expect("write touched source");
        if commit_everything {
            git(&root, &["add", "."]);
            commit(&root, "second");
        }
        root
    }

    fn commit(root: &PathBuf, message: &str) {
        git(
            root,
            &[
                "-c",
                "user.name=test",
                "-c",
                "user.email=test@example.com",
                "commit",
                "-qm",
                message,
            ],
        );
    }

    fn git(root: &PathBuf, args: &[&str]) {
        // Ignore the developer's global and system git config. A global
        // core.hooksPath can drop generated files into the fixture on commit,
        // which then show up as untracked changes and skew the assertions.
        let output = std::process::Command::new("git")
            .args(args)
            .current_dir(root)
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .output()
            .expect("run git");
        assert!(
            output.status.success(),
            "git {}: {}",
            args.join(" "),
            String::from_utf8_lossy(&output.stderr)
        );
    }

    fn write_fixture_project() -> PathBuf {
        let root = new_temp_dir("cli");
        let src_dir = root.join("src");
        let gauntlet_dir = root.join(".gauntlet");
        fs::create_dir_all(&src_dir).expect("create src directory");
        fs::create_dir_all(&gauntlet_dir).expect("create .gauntlet directory");

        let source = r#"
pub fn hot(v: i32) -> i32 {
    if v > 0 { 1 } else { 0 }
}

pub fn cold() -> i32 {
    0
}
"#;
        fs::write(src_dir.join("lib.rs"), source).expect("write source");
        fs::write(
            gauntlet_dir.join("thresholds.yml"),
            "metrics:\n  crap_ceiling:\n    direction: max\n    value: 2\n",
        )
        .expect("write thresholds");

        let absolute_file = src_dir.join("lib.rs").to_string_lossy().to_string();
        let coverage = format!(
            "{{\"data\":[{{\"functions\":[{{\"name\":\"sample::hot\",\"count\":1,\"filenames\":[\"{absolute_file}\"],\"regions\":[[2,1,2,30,1,0]]}}]}}]}}"
        );
        fs::write(root.join("coverage.json"), coverage).expect("write coverage");
        fs::write(
            root.join("coverage-none.json"),
            "{\"data\":[{\"functions\":[]}]}",
        )
        .expect("write empty coverage");

        root
    }

    fn bytes_to_string(buffer: &[u8]) -> String {
        String::from_utf8_lossy(buffer).to_string()
    }

    // The clock alone is not a unique name. `cargo test` runs these in parallel
    // threads and SystemTime here is only microsecond-resolution, so two
    // fixtures really do land on the same nanos value and the second one then
    // reuses the first's half-built directory. The counter makes the name unique
    // per call; the pid keeps concurrent `cargo test` runs apart.
    fn new_temp_dir(prefix: &str) -> PathBuf {
        static COUNTER: AtomicU64 = AtomicU64::new(0);
        let nanos = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .expect("clock before epoch")
            .as_nanos();
        let ordinal = COUNTER.fetch_add(1, Ordering::Relaxed);
        let pid = std::process::id();
        let dir = std::env::temp_dir().join(format!("crap4rs-{prefix}-{pid}-{ordinal}-{nanos}"));
        fs::create_dir_all(&dir).expect("create temp dir");
        dir
    }
}
