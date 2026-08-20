export function buildKey(file: string, line: number, func: string): string {
  return `${file}:${line}:${func}`;
}
