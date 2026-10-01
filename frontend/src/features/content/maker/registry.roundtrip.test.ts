import { describe, it, expect } from 'vitest';
import { levelTemplateRegistry } from './registry';
import type { TemplateType } from '../types';

// plan/05_Contenido_maker_y_juego.md §7.3: "round-trip cerrado del maker...
// se prueba con un test por cada una de las siete plantillas". Responde a la
// pregunta concreta que motivó F10.9: ¿el exportador conserva fríamente el
// mismo formato que acepta la entrada?
//
// getDefaults() NO sirve como fixture aquí: varias plantillas arrancan con
// campos requeridos vacíos a propósito (p. ej. `question: ''` en trivia) —
// es el estado de un formulario en blanco, no un nivel válido. Cada fixture
// de abajo es lo que produciría un formulario YA LLENO correctamente.
const VALID_CONTENT_FIXTURES: Record<TemplateType, unknown> = {
  trivia: [
    { question: '¿Capital de México?', options: ['Ciudad de México', 'Guadalajara'], correct_index: 0 },
    { question: '¿Capital de Francia?', options: ['París', 'Madrid'], correct_index: 0 },
    { question: '¿Capital de Japón?', options: ['Tokio', 'Seúl'], correct_index: 0 },
  ],
  crossword: {
    words: [
      { word: 'GATO', clue: 'Felino doméstico' },
      { word: 'GALLO', clue: 'Ave de corral que canta al amanecer' },
    ],
  },
  word_search: { words: ['SOL', 'LUNA'], width: 12, height: 12, seed: 1234 },
  puzzle: { phrase: 'Universidad Veracruzana', pieces: 4, seed: 1234 },
  fake_news: {
    news: [
      { title: 'Titular de prueba', content: 'Contenido de prueba', isFake: false, reference: 'https://example.com' },
    ],
  },
  memory: {
    back_color: '#18529D',
    pairs: [
      { id: '1', content1: 'A', content2: 'A' },
      { id: '2', content1: 'B', content2: 'B' },
      { id: '3', content1: 'C', content2: 'C' },
      { id: '4', content1: 'D', content2: 'D' },
    ],
  },
  snakes_ladders: {
    board_width: 6,
    board_height: 6,
    start_position: 1,
    end_position: 36,
    seed: 2026,
    snakes: [{ start: 20, end: 5 }],
    ladders: [{ start: 3, end: 15 }],
    ai_config: { difficulty: 'MEDIUM' },
    questions: Array.from({ length: 8 }, (_, i) => ({
      question: `Pregunta ${i + 1}`,
      options: ['A', 'B'],
      correct_index: 0,
    })),
  },
};

const TEMPLATE_TYPES = Object.keys(levelTemplateRegistry) as TemplateType[];

describe('levelTemplateRegistry — round-trip de exportación por plantilla', () => {
  it('cubre las siete plantillas del vocabulario oficial', () => {
    expect(TEMPLATE_TYPES.sort()).toEqual(
      ['crossword', 'fake_news', 'memory', 'puzzle', 'snakes_ladders', 'trivia', 'word_search'].sort()
    );
  });

  it.each(TEMPLATE_TYPES)('%s: getDefaults() tiene la forma correcta para su propio FormComponent', (type) => {
    // No exigimos que pase el schema (varias plantillas arrancan con campos
    // requeridos vacíos a propósito), solo que no explote al validarse.
    const entry = levelTemplateRegistry[type];
    expect(() => entry.schema.safeParse(entry.getDefaults())).not.toThrow();
  });

  it.each(TEMPLATE_TYPES)('%s: un nivel lleno correctamente pasa su propio schema', (type) => {
    const entry = levelTemplateRegistry[type];
    const result = entry.schema.safeParse(VALID_CONTENT_FIXTURES[type]);
    expect(result.success).toBe(true);
  });

  it.each(TEMPLATE_TYPES)('%s: sobrevive intacto un ciclo JSON.stringify → JSON.parse', (type) => {
    const entry = levelTemplateRegistry[type];
    const original = VALID_CONTENT_FIXTURES[type];

    const roundTripped: unknown = JSON.parse(JSON.stringify(original));

    expect(roundTripped).toEqual(original);
    expect(entry.schema.safeParse(roundTripped).success).toBe(true);
  });

  it.each(TEMPLATE_TYPES)('%s: el objeto completo del maker (metadata + content) sobrevive el ciclo JSON', (type) => {
    const entry = levelTemplateRegistry[type];
    const content = VALID_CONTENT_FIXTURES[type];
    const exported = {
      metadata: {
        id: '00000000-0000-0000-0000-000000000000',
        title: 'Nivel de prueba',
        author: 'Autor de prueba',
        color: '#18529D',
        difficulty: 1,
        template_type: type,
        creation_date: '2026-08-26T00:00:00.000Z',
      },
      content,
    };

    const roundTripped: typeof exported = JSON.parse(JSON.stringify(exported));

    expect(roundTripped).toEqual(exported);
    expect(entry.schema.safeParse(roundTripped.content).success).toBe(true);
  });
});
