use std::cmp::Ordering;
use std::io::Write;
use std::path::{Path, PathBuf};

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
    _stderr: &mut dyn Write,
) -> Result<i32, String> {
    let root = std::fs::canonicalize(&options.dir)
        .map_err(|error| format!("resolving scan root {}: {error}", options.dir.display()))?;
    let coverage_json_path = resolve_path(&options.coverage_json, &root);
    let ceiling_file_path = resolve_path(&options.ceiling_file, &root);
    let mut functions = scan_dir(&root)?;
    functions.sort_by(sort_complexity_records);

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
    let mut coverage_json = PathBuf::from("coverage.json");
    let mut ceiling_file = PathBuf::from(".gauntlet/thresholds.yml");
    let mut ceiling_override = None;
    let mut format = OutputFormat::Text;

    let mut index = 0;
    while index < args.len() {
        match args[index].as_str() {
            "--dir" => {
                index += 1;
                dir = parse_path_arg(args, index, "--dir")?;
            }
            "--coverage-json" => {
                index += 1;
                coverage_json = parse_path_arg(args, index, "--coverage-json")?;
            }
            "--ceiling-file" => {
                index += 1;
                ceiling_file = parse_path_arg(args, index, "--ceiling-file")?;
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
    "usage: crap4rs [--dir <dir>] [--coverage-json <path>] [--ceiling-file <path>] [--ceiling <n>] [--format text|json]"
}

#[cfg(test)]
mod tests {
    use std::fs;
    use std::io::Cursor;
    use std::path::PathBuf;
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

    fn new_temp_dir(prefix: &str) -> PathBuf {
        let suffix = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .expect("clock before epoch")
            .as_nanos();
        let dir = std::env::temp_dir().join(format!("crap4rs-{prefix}-{suffix}"));
        fs::create_dir_all(&dir).expect("create temp dir");
        dir
    }
}
