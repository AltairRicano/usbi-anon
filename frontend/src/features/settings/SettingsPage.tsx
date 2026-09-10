import { HomeButton } from '../../shared/components/ui/HomeButton';
import {
  useSettingsStore,
  type ColorBlindFilter,
  type TextScale,
} from './useSettingsStore';

// Ruta pública a propósito (App.tsx no la envuelve en ProtectedRoute):
// el registro de este proyecto es un cuestionario de 5 preguntas de texto
// libre, así que alguien con baja visión o daltonismo necesita poder ajustar
// tema/tamaño de texto/filtro de color ANTES de tener cuenta, no solo
// después de iniciar sesión.
export default function SettingsPage() {
  const {
    theme,
    setTheme,
    colorBlindFilter,
    setColorBlindFilter,
    muteGameSounds,
    setMuteGameSounds,
    reduceMotion,
    setReduceMotion,
    textScale,
    setTextScale,
  } = useSettingsStore();

  return (
    <main className="min-h-screen p-6" style={{ backgroundColor: 'var(--color-surface)' }}>
      <div className="mx-auto max-w-3xl space-y-6">
        <header className="flex flex-wrap items-center justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold">Configuración</h1>
            <p className="text-sm text-[--color-muted]">Ajusta la apariencia y accesibilidad de la aplicación.</p>
          </div>
          <HomeButton />
        </header>

        <section className="rounded-lg bg-[--color-card] p-5 shadow-sm">
          <h2 className="mb-1 text-xl font-semibold">Apariencia</h2>
          <p className="mb-4 text-sm text-[--color-muted]">Tema de color y tamaño de texto.</p>

          <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
            <fieldset>
              <legend className="mb-2 text-sm font-semibold">Tema</legend>
              <div className="flex gap-2" role="radiogroup" aria-label="Tema">
                <SegmentButton active={theme === 'light'} onClick={() => setTheme('light')}>
                  Claro
                </SegmentButton>
                <SegmentButton active={theme === 'dark'} onClick={() => setTheme('dark')}>
                  Oscuro
                </SegmentButton>
              </div>
            </fieldset>

            <fieldset>
              <legend className="mb-2 text-sm font-semibold">Tamaño de texto</legend>
              <div className="flex gap-2" role="radiogroup" aria-label="Tamaño de texto">
                <SegmentButton active={textScale === 'normal'} onClick={() => setTextScale('normal' as TextScale)}>
                  Normal
                </SegmentButton>
                <SegmentButton active={textScale === 'large'} onClick={() => setTextScale('large' as TextScale)}>
                  Grande
                </SegmentButton>
                <SegmentButton active={textScale === 'larger'} onClick={() => setTextScale('larger' as TextScale)}>
                  Más grande
                </SegmentButton>
              </div>
            </fieldset>
          </div>
        </section>

        <section className="rounded-lg bg-[--color-card] p-5 shadow-sm">
          <h2 className="mb-1 text-xl font-semibold">Accesibilidad visual</h2>
          <p className="mb-4 text-sm text-[--color-muted]">
            Filtro de daltonización: redistribuye el color hacia tonos que sí distingas (no es un simulador).
          </p>
          <label className="block max-w-sm">
            <span className="sr-only">Filtro de daltonización</span>
            <select
              value={colorBlindFilter}
              onChange={(e) => setColorBlindFilter(e.target.value as ColorBlindFilter)}
              className="w-full rounded border border-gray-300 bg-white px-3 py-2 text-sm text-black focus:outline-none focus-visible:ring-2"
            >
              <option value="none">Normal (sin filtro)</option>
              <option value="deuteranopia">Deuteranopía (Verde-Rojo)</option>
              <option value="protanopia">Protanopía (Rojo-Verde)</option>
              <option value="tritanopia">Tritanopía (Azul-Amarillo)</option>
            </select>
          </label>
        </section>

        <section className="rounded-lg bg-[--color-card] p-5 shadow-sm">
          <h2 className="mb-1 text-xl font-semibold">Movimiento</h2>
          <p className="mb-4 text-sm text-[--color-muted]">
            La aplicación ya respeta la preferencia "reducir movimiento" de tu sistema operativo. Actívalo aquí
            si tu dispositivo no expone esa opción o prefieres anularla manualmente.
          </p>
          <ToggleRow
            label="Reducir animaciones"
            checked={reduceMotion}
            onChange={setReduceMotion}
          />
        </section>

        <section className="rounded-lg bg-[--color-card] p-5 shadow-sm">
          <h2 className="mb-1 text-xl font-semibold">Sonido</h2>
          <p className="mb-4 text-sm text-[--color-muted]">
            Controla el sonido de los minijuegos (efectos de dados, temporizadores, etc.).
          </p>
          <ToggleRow
            label="Silenciar sonidos de juego"
            checked={muteGameSounds}
            onChange={setMuteGameSounds}
          />
        </section>
      </div>
    </main>
  );
}

function SegmentButton({
  active,
  onClick,
  children,
}: {
  active: boolean;
  onClick: () => void;
  children: React.ReactNode;
}) {
  return (
    <button
      type="button"
      role="radio"
      aria-checked={active}
      onClick={onClick}
      className="rounded-lg border px-4 py-2 text-sm font-medium transition-colors"
      style={
        active
          ? { backgroundColor: 'var(--color-primary)', color: 'var(--color-primary-foreground)', borderColor: 'var(--color-primary)' }
          : { backgroundColor: 'transparent', color: 'var(--color-foreground)', borderColor: 'var(--color-border)' }
      }
    >
      {children}
    </button>
  );
}

function ToggleRow({
  label,
  checked,
  onChange,
}: {
  label: string;
  checked: boolean;
  onChange: (value: boolean) => void;
}) {
  return (
    <label className="flex min-h-touch cursor-pointer items-center gap-3">
      <input
        type="checkbox"
        className="min-h-touch min-w-touch"
        checked={checked}
        onChange={(e) => onChange(e.target.checked)}
      />
      <span className="text-sm">{label}</span>
    </label>
  );
}
