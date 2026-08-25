import axios from 'axios';

/**
 * Traduce un error de red a un mensaje legible, priorizando el `detail` de
 * un RFC 7807 Problem Details (ver httpproblem.WriteProblem en Go) sobre el
 * mensaje genérico de axios. Extraído a shared/ (plan/04 §5) porque las tres
 * páginas de esta fase (login, banco de preguntas, admin de cuentas) repiten
 * el mismo manejo — en ../usbi vivía duplicado inline en cada página.
 */
export function errorMessage(err: unknown, fallback: string): string {
  if (axios.isAxiosError(err)) {
    const detail = (err.response?.data as { detail?: string } | undefined)?.detail;
    return detail ?? fallback;
  }
  if (err instanceof Error) {
    return err.message;
  }
  return fallback;
}
