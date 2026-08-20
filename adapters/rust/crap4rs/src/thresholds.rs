use std::fs;
use std::path::Path;

pub fn read_crap_ceiling(path: &Path) -> Result<f64, String> {
    let source = fs::read_to_string(path)
        .map_err(|error| format!("reading thresholds file {}: {error}", path.display()))?;
    parse_crap_ceiling(&source).ok_or_else(|| {
        format!(
            "missing metrics.crap_ceiling.value in thresholds file {}",
            path.display()
        )
    })
}

fn parse_crap_ceiling(source: &str) -> Option<f64> {
    let mut in_metrics = false;
    let mut in_crap_ceiling = false;

    for raw_line in source.lines() {
        let line_without_comment = raw_line.split('#').next().unwrap_or("");
        if line_without_comment.trim().is_empty() {
            continue;
        }

        let indent = raw_line
            .chars()
            .take_while(|ch| ch.is_ascii_whitespace())
            .count();
        let trimmed = line_without_comment.trim();

        if indent == 0 {
            in_metrics = trimmed == "metrics:";
            in_crap_ceiling = false;
            continue;
        }

        if !in_metrics {
            continue;
        }

        if indent == 2 && trimmed.ends_with(':') {
            in_crap_ceiling = trimmed == "crap_ceiling:";
            continue;
        }

        if in_crap_ceiling && indent >= 4 && trimmed.starts_with("value:") {
            let value = trimmed.trim_start_matches("value:").trim();
            return value.parse::<f64>().ok();
        }
    }

    None
}

#[cfg(test)]
mod tests {
    use super::parse_crap_ceiling;

    #[test]
    fn parses_nested_thresholds_value() {
        let source = r#"
metrics:
  mutation_mcover:
    direction: min
    value: 88
  crap_ceiling:
    direction: max
    value: 30
"#;

        assert_eq!(parse_crap_ceiling(source), Some(30.0));
    }
}
