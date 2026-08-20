import fs from "node:fs";

import yaml from "js-yaml";

interface ThresholdsFile {
  metrics?: {
    crap_ceiling?: {
      direction?: string;
      value?: number;
    };
  };
}

export function readCeiling(thresholdsPath: string): number {
  const raw = fs.readFileSync(thresholdsPath, "utf8");
  const parsed = yaml.load(raw) as ThresholdsFile;
  const value = parsed?.metrics?.crap_ceiling?.value;
  if (typeof value !== "number" || !Number.isFinite(value)) {
    throw new Error(`thresholds file ${thresholdsPath} is missing metrics.crap_ceiling.value`);
  }
  return value;
}
