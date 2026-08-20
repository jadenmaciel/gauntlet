#[derive(Clone, Debug, PartialEq)]
pub struct FunctionComplexity {
    pub file: String,
    pub line: usize,
    pub func_name: String,
    pub complexity: u64,
}

#[derive(Clone, Debug, PartialEq)]
pub struct FunctionReport {
    pub file: String,
    pub line: usize,
    pub func_name: String,
    pub complexity: u64,
    pub coverage: f64,
    pub crap: f64,
    pub pass: bool,
    pub matched: bool,
}

#[derive(Clone, Debug, PartialEq)]
pub struct Summary {
    pub total: usize,
    pub failing: usize,
    pub max_crap: f64,
}

#[derive(Clone, Debug, PartialEq)]
pub struct Report {
    pub ceiling: f64,
    pub functions: Vec<FunctionReport>,
    pub summary: Summary,
}

pub fn build_key(file: &str, line: usize, func_name: &str) -> String {
    format!("{file}:{line}:{func_name}")
}
