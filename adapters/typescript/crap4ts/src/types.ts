export interface FunctionComplexity {
  file: string;
  line: number;
  func: string;
  complexity: number;
}

export interface FunctionReport {
  file: string;
  line: number;
  func: string;
  complexity: number;
  coverage: number;
  crap: number;
  pass: boolean;
  matched: boolean;
}

export interface Summary {
  total: number;
  failing: number;
  max_crap: number;
}

export interface Report {
  ceiling: number;
  functions: FunctionReport[];
  summary: Summary;
}

export interface CliFunctionReport {
  file: string;
  line: number;
  func: string;
  complexity: number;
  coverage: number;
  crap: number;
  pass: boolean;
}

export interface CliReport {
  ceiling: number;
  functions: CliFunctionReport[];
  summary: Summary;
}
