use std::collections::HashMap;

use crate::model::{build_key, FunctionComplexity, FunctionReport, Report, Summary};
use crate::score::compute_crap;

pub fn evaluate(
    functions: &[FunctionComplexity],
    coverage_percent: &HashMap<String, f64>,
    ceiling: f64,
) -> Report {
    let mut reports = Vec::with_capacity(functions.len());
    let mut failing = 0_usize;
    let mut max_crap = 0.0_f64;

    for function in functions {
        let key = build_key(&function.file, function.line, &function.func_name);
        let matched = coverage_percent.contains_key(&key);
        let coverage = coverage_percent.get(&key).copied().unwrap_or(0.0);
        let crap = compute_crap(function.complexity, coverage);
        let pass = crap <= ceiling;
        if !pass {
            failing += 1;
        }
        if crap > max_crap {
            max_crap = crap;
        }

        reports.push(FunctionReport {
            file: function.file.clone(),
            line: function.line,
            func_name: function.func_name.clone(),
            complexity: function.complexity,
            coverage,
            crap,
            pass,
            matched,
        });
    }

    Report {
        ceiling,
        summary: Summary {
            total: reports.len(),
            failing,
            max_crap,
        },
        functions: reports,
    }
}

#[cfg(test)]
mod tests {
    use std::collections::HashMap;

    use super::evaluate;
    use crate::model::{build_key, FunctionComplexity};

    #[test]
    fn uncovered_functions_are_scored_and_marked_unmatched() {
        let functions = vec![
            FunctionComplexity {
                file: "src/lib.rs".to_string(),
                line: 10,
                func_name: "covered".to_string(),
                complexity: 5,
            },
            FunctionComplexity {
                file: "src/lib.rs".to_string(),
                line: 30,
                func_name: "uncovered".to_string(),
                complexity: 5,
            },
        ];

        let mut coverage = HashMap::new();
        coverage.insert(build_key("src/lib.rs", 10, "covered"), 50.0);

        let report = evaluate(&functions, &coverage, 20.0);
        assert_eq!(report.summary.total, 2);

        let covered = &report.functions[0];
        assert!(covered.matched);
        assert!((covered.coverage - 50.0).abs() < 1e-9);

        let uncovered = &report.functions[1];
        assert!(!uncovered.matched);
        assert!((uncovered.coverage - 0.0).abs() < 1e-9);
        assert!((uncovered.crap - 30.0).abs() < 1e-9);
    }
}
