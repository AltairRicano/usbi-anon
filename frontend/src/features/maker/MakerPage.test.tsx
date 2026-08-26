import { render, screen, fireEvent, waitFor, cleanup } from '@testing-library/react';
import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import { MakerPage } from './MakerPage';

// F10.8 (plan/05_Contenido_maker_y_juego.md §5): el archivo de ../usbi tenía
// mocks de Tauri y aserciones (título "Maker — Creación de Niveles", botón
// "Exportar a Local", placeholders "¿Cuál es...?"/"Opción 1"-4) que ya no
// coincidían con su propio MakerPage.tsx ni con TriviaForm.tsx en ningún
// punto de este porteo — no es una regresión de esta fase, el original ya
// estaba desalineado consigo mismo. Reescrito contra el DOM real.

describe('MakerPage', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  // Sin test.globals: true en vite.config.ts, @testing-library/react no
  // encuentra un afterEach global del que colgar su cleanup automático —
  // sin esto, cada `render()` de un test deja su DOM montado para el
  // siguiente y los queries por rol empiezan a matchear duplicados.
  afterEach(cleanup);

  it('renders the metadata form with correct labels', () => {
    render(
      <MemoryRouter>
        <MakerPage />
      </MemoryRouter>
    );

    expect(screen.getByText('Maker Local — Creación de Niveles')).toBeDefined();
    expect(screen.getByLabelText('Título')).toBeDefined();
    expect(screen.getByLabelText('Autor')).toBeDefined();
    expect(screen.getByLabelText('Dificultad (1–10)')).toBeDefined();
    expect(screen.getByLabelText('Tipo de Plantilla')).toBeDefined();
    expect(screen.getByRole('button', { name: 'Crear (Guardar en Local)' })).toBeDefined();
    expect(screen.getByRole('button', { name: 'Exportar (Descargar JSON)' })).toBeDefined();
  });

  it('renders the Trivia sub-editor by default, with 3 questions', () => {
    render(
      <MemoryRouter>
        <MakerPage />
      </MemoryRouter>
    );

    expect(screen.getByLabelText('Pregunta 1')).toBeDefined();
    expect(screen.getByLabelText('Pregunta 2')).toBeDefined();
    expect(screen.getByLabelText('Pregunta 3')).toBeDefined();
  });

  it('refuses to save when the title is empty', async () => {
    render(
      <MemoryRouter>
        <MakerPage />
      </MemoryRouter>
    );

    fireEvent.click(screen.getByRole('button', { name: 'Crear (Guardar en Local)' }));

    await waitFor(() => {
      expect(screen.getByText(/Revisa que el título no esté vacío/i)).toBeDefined();
    });
    expect(localStorage.getItem('usbi_local_levels')).toBeNull();
  });

  it('saves to localStorage under usbi_local_levels when the form is valid', async () => {
    render(
      <MemoryRouter>
        <MakerPage />
      </MemoryRouter>
    );

    fireEvent.change(screen.getByLabelText('Título'), { target: { value: 'Mi Nivel de Prueba' } });
    fireEvent.change(screen.getByLabelText('Autor'), { target: { value: 'Profe UV' } });

    fireEvent.change(screen.getByLabelText('Pregunta 1'), { target: { value: '¿Capital de México?' } });
    fireEvent.change(screen.getByLabelText('Pregunta 2'), { target: { value: '¿Capital de Francia?' } });
    fireEvent.change(screen.getByLabelText('Pregunta 3'), { target: { value: '¿Capital de Japón?' } });

    fireEvent.click(screen.getByRole('button', { name: 'Crear (Guardar en Local)' }));

    await waitFor(() => {
      expect(screen.getByText(/Operación realizada exitosamente/i)).toBeDefined();
    });

    const stored = JSON.parse(localStorage.getItem('usbi_local_levels') ?? '[]');
    expect(stored).toHaveLength(1);
    expect(stored[0].metadata.title).toBe('Mi Nivel de Prueba');
    expect(stored[0].metadata.author).toBe('Profe UV');
    expect(stored[0].metadata.template_type).toBe('trivia');
    expect(stored[0].content[0].question).toBe('¿Capital de México?');
  });
});
