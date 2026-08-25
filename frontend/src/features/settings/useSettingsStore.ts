import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';

export type ColorBlindFilter = 'none' | 'deuteranopia' | 'protanopia' | 'tritanopia';
export type TextScale = 'normal' | 'large' | 'larger';

const TEXT_SCALE_CLASSES = ['text-scale-large', 'text-scale-larger'];

interface SettingsState {
  colorBlindFilter: ColorBlindFilter;
  setColorBlindFilter: (filter: ColorBlindFilter) => void;
  theme: 'light' | 'dark';
  setTheme: (theme: 'light' | 'dark') => void;
  /**
   * Portado de ../usbi para cuando lleguen los minijuegos (no forman parte
   * del alcance de F10) — el campo vive en el store desde ya (costo cero,
   * no toca backend) pero SettingsPage todavía no renderiza un control para
   * él. Activar su sección ahí el día que exista un minijuego con sonido.
   */
  muteGameSounds: boolean;
  setMuteGameSounds: (muted: boolean) => void;
  /**
   * Override explícito sobre la media query `prefers-reduced-motion` del SO
   * (que index.css ya respeta automáticamente sin importar este flag) — para
   * quien quiere movimiento reducido pero su SO/navegador no expone la
   * preferencia con facilidad.
   */
  reduceMotion: boolean;
  setReduceMotion: (reduce: boolean) => void;
  textScale: TextScale;
  setTextScale: (scale: TextScale) => void;
}

export function applyDocumentClasses(state: Pick<SettingsState, 'colorBlindFilter' | 'theme' | 'reduceMotion' | 'textScale'>) {
  const html = document.documentElement;

  html.classList.toggle('dark', state.theme === 'dark');

  html.classList.remove('deuteranopia', 'protanopia', 'tritanopia');
  if (state.colorBlindFilter !== 'none') {
    html.classList.add(state.colorBlindFilter);
  }

  html.classList.toggle('reduce-motion', state.reduceMotion);

  html.classList.remove(...TEXT_SCALE_CLASSES);
  if (state.textScale === 'large') html.classList.add('text-scale-large');
  if (state.textScale === 'larger') html.classList.add('text-scale-larger');
}

// Preferencias de apariencia/accesibilidad: deliberadamente en localStorage,
// a diferencia de useAuthStore (sessionStorage). El token de sesión debe
// morir al cerrar el navegador; la preferencia de alguien con baja visión o
// daltonismo no — y esta ruta es pública precisamente para que se pueda
// configurar antes de tener cuenta (ver RegisterPage/LoginPage).
export const useSettingsStore = create<SettingsState>()(
  persist(
    (set, get) => ({
      colorBlindFilter: 'none',
      theme: 'light',
      muteGameSounds: false,
      reduceMotion: false,
      textScale: 'normal',
      setColorBlindFilter: (filter: ColorBlindFilter) => {
        set({ colorBlindFilter: filter });
        applyDocumentClasses(get());
      },
      setTheme: (theme: 'light' | 'dark') => {
        set({ theme });
        applyDocumentClasses(get());
      },
      setMuteGameSounds: (muted: boolean) => set({ muteGameSounds: muted }),
      setReduceMotion: (reduce: boolean) => {
        set({ reduceMotion: reduce });
        applyDocumentClasses(get());
      },
      setTextScale: (scale: TextScale) => {
        set({ textScale: scale });
        applyDocumentClasses(get());
      },
    }),
    {
      name: 'usbi-anon-settings',
      storage: createJSONStorage(() => localStorage),
      onRehydrateStorage: () => (state) => {
        // Al recargar la página, restaurar las clases en el HTML.
        if (state) {
          applyDocumentClasses(state);
        }
      },
    }
  )
);
