// F10.8 (plan/05_Contenido_maker_y_juego.md §5): el original renderiza el
// PuzzleGame real (features/games/PuzzleGame.tsx) dentro de la
// previsualización, pero features/games/ todavía no se ha portado — eso es
// F10.10. Placeholder temporal: se sustituye por el import real de
// PuzzleGame en cuanto F10.10 traiga la carpeta games/.
export function PuzzlePreview({ value }: { value: { phrase?: string; pieces?: number; seed?: number } }) {
  if (!value.phrase) {
    return (
      <div className="p-4 bg-gray-50 border rounded-md text-[--color-muted] text-center">
        Ingresa una frase para ver la vista previa.
      </div>
    );
  }

  return (
    <div className="p-4 bg-gray-50 border rounded-md text-center text-[--color-muted]">
      <h3 className="font-bold text-gray-700 mb-2">Vista previa</h3>
      <p className="text-sm">
        «{value.phrase}» — {value.pieces ?? 3} piezas. La previsualización jugable del rompecabezas
        estará disponible cuando se porten los minijuegos (F10.10).
      </p>
    </div>
  );
}
