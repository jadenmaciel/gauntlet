pub fn compute_crap(complexity: u64, coverage_percent: f64) -> f64 {
    let cc = complexity as f64;
    let coverage = coverage_percent / 100.0;
    let uncovered = 1.0 - coverage;
    cc * cc * uncovered * uncovered * uncovered + cc
}

#[cfg(test)]
mod tests {
    use super::compute_crap;

    #[test]
    fn matches_golden_vectors() {
        let epsilon = 1e-9;
        let vectors = [
            (1_u64, 0.0_f64, 2.0_f64),
            (1, 100.0, 1.0),
            (2, 0.0, 6.0),
            (3, 50.0, 4.125),
            (5, 0.0, 30.0),
            (5, 50.0, 8.125),
            (5, 100.0, 5.0),
            (8, 100.0, 8.0),
            (10, 0.0, 110.0),
            (10, 80.0, 10.8),
            (10, 100.0, 10.0),
            (15, 100.0, 15.0),
            (20, 90.0, 20.4),
        ];

        for (complexity, coverage, expected) in vectors {
            let actual = compute_crap(complexity, coverage);
            assert!(
                (actual - expected).abs() < epsilon,
                "vector mismatch for cc={complexity}, coverage={coverage}: expected {expected}, got {actual}",
            );
        }
    }
}
