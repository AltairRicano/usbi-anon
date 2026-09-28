export function weightedRandom<T>(options: T[], weights: number[], random: () => number = Math.random): T {
  if (options.length === 0 || options.length !== weights.length) {
    throw new Error("Invalid options or weights");
  }
  let sum = 0;
  for (const w of weights) {
    if (w < 0) throw new Error("Negative weight not allowed");
    sum += w;
  }
  if (sum === 0) throw new Error("Total weight cannot be zero");

  let r = random() * sum;
  for (let i = 0; i < options.length; i++) {
    r -= weights[i];
    if (r <= 0) return options[i];
  }
  return options[options.length - 1];
}
