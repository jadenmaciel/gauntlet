export function hotPath(input: number): number {
  if (input > 10) return 1;
  if (input > 9) return 2;
  if (input > 8) return 3;
  if (input > 7) return 4;
  if (input > 6) return 5;
  if (input > 5) return 6;
  if (input > 4) return 7;
  if (input > 3) return 8;
  if (input > 2) return 9;
  return 10;
}
