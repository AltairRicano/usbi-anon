export function mapToGrid(index: number, width: number, height: number): { x: number, y: number } {
  if (index < 1 || index > width * height) throw new Error("Index out of bounds");
  
  // 1-based to 0-based
  const i = index - 1;
  const row = Math.floor(i / width);
  const isEvenRow = row % 2 === 0;
  
  // Bottom-up rows (row 0 is bottom)
  const y = height - 1 - row;
  
  // Alternating columns
  let x = i % width;
  if (!isEvenRow) {
    x = width - 1 - x;
  }
  
  return { x, y };
}
