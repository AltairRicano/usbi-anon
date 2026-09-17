import { useEffect, useState, type FormEvent } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { ZodError } from 'zod';
import { Button } from '../../shared/components/ui/Button';
import { Input } from '../../shared/components/ui/Input';
import { UsbiEmblem, Spinner } from '../../shared/components/ui/Brand';
import { SettingsEntry } from '../../shared/components/SettingsEntry';
import { apiClient } from '../../shared/apiClient';
import { errorMessage } from '../../shared/errorMessage';
import { usePrivacyNotice } from '../legal/usePrivacyNotice';
import { PrivacyNoticeInline } from '../legal/PrivacyNoticeInline';
import {
  AnswerTextSchema,
  RegisterAnswersResponseSchema,
  RegisterConfirmResponseSchema,
  RegisterQuestionsResponseSchema,
} from './schemas';

type Step = 1 | 2 | 3;

interface QuestionOption {
  id: string;
  text: string;
}

export default function RegisterPage() {
  const navigate = useNavigate();

  const [step, setStep] = useState<Step>(1);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // ── Paso 1: cuestionario de gustos ─────────────────────────────────────────
  const [questions, setQuestions] = useState<QuestionOption[]>([]);
  const [answers, setAnswers] = useState<Record<string, string>>({});
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [isAdult, setIsAdult] = useState(false);
  const [acceptedPrivacy, setAcceptedPrivacy] = useState(false);
  const [loadingQuestions, setLoadingQuestions] = useState(true);
  const { notice: privacyNotice, loading: loadingPrivacyNotice, error: privacyNoticeError } = usePrivacyNotice();

  // ── Paso 2: elegir nickname ─────────────────────────────────────────────
  const [registrationToken, setRegistrationToken] = useState('');
  const [nicknameCandidates, setNicknameCandidates] = useState<string[]>([]);
  const [chosenNickname, setChosenNickname] = useState('');

  // ── Paso 3: credencial emitida (se muestra UNA sola vez) ──────────────────
  const [issuedNickname, setIssuedNickname] = useState('');
  const [issuedPassword, setIssuedPassword] = useState('');
  const [displayAlias, setDisplayAlias] = useState('');
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    let cancelled = false;
    async function loadQuestions() {
      setLoadingQuestions(true);
      setError(null);
      try {
        const resp = await apiClient.post('/auth/register/questions');
        const data = RegisterQuestionsResponseSchema.parse(resp.data);
        if (!cancelled) setQuestions(data.questions);
      } catch (err) {
        if (!cancelled) setError(errorMessage(err, 'No se pudieron cargar las preguntas de registro.'));
      } finally {
        if (!cancelled) setLoadingQuestions(false);
      }
    }
    void loadQuestions();
    return () => {
      cancelled = true;
    };
  }, []);

  async function handleAnswersSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError(null);

    const nextFieldErrors: Record<string, string> = {};
    const payloadAnswers: { question_id: string; answer_text: string }[] = [];
    for (const q of questions) {
      const raw = answers[q.id] ?? '';
      const parsed = AnswerTextSchema.safeParse(raw);
      if (!parsed.success) {
        nextFieldErrors[q.id] = parsed.error.issues[0]?.message ?? 'Respuesta inválida.';
        continue;
      }
      payloadAnswers.push({ question_id: q.id, answer_text: parsed.data });
    }
    setFieldErrors(nextFieldErrors);
    if (Object.keys(nextFieldErrors).length > 0) return;
    if (!acceptedPrivacy) {
      setError('Debes aceptar el aviso de privacidad para continuar.');
      return;
    }
    if (!privacyNotice) {
      setError('No se pudo cargar el aviso de privacidad. Recarga la página e intenta de nuevo.');
      return;
    }

    setLoading(true);
    try {
      const resp = await apiClient.post('/auth/register/answers', {
        answers: payloadAnswers,
        is_adult: isAdult,
        privacy_notice_version: privacyNotice.version,
      });
      const data = RegisterAnswersResponseSchema.parse(resp.data);
      setRegistrationToken(data.registration_token);
      setNicknameCandidates(data.nickname_candidates);
      setChosenNickname(data.nickname_candidates[0] ?? '');
      setStep(2);
    } catch (err) {
      if (err instanceof ZodError) {
        console.error('[RegisterPage] Respuesta de /auth/register/answers con forma inesperada:', err.issues);
        setError('El servidor respondió de forma inesperada. Intenta de nuevo más tarde.');
      } else {
        setError(errorMessage(err, 'No se pudieron validar tus respuestas.'));
      }
    } finally {
      setLoading(false);
    }
  }

  async function handleConfirmSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      const resp = await apiClient.post('/auth/register/confirm', {
        registration_token: registrationToken,
        chosen_nickname: chosenNickname,
      });
      const data = RegisterConfirmResponseSchema.parse(resp.data);
      setIssuedNickname(data.nickname);
      setIssuedPassword(data.password);
      setDisplayAlias(data.display_alias);
      setStep(3);
    } catch (err) {
      if (err instanceof ZodError) {
        console.error('[RegisterPage] Respuesta de /auth/register/confirm con forma inesperada:', err.issues);
        setError('El servidor respondió de forma inesperada. Intenta de nuevo más tarde.');
      } else {
        setError(errorMessage(err, 'No se pudo confirmar el registro. El token pudo expirar (10 minutos): vuelve a intentarlo desde el paso 1.'));
      }
    } finally {
      setLoading(false);
    }
  }

  async function handleCopyCredentials() {
    try {
      await navigator.clipboard.writeText(`Nickname: ${issuedNickname}\nContraseña: ${issuedPassword}`);
      setCopied(true);
    } catch {
      // Clipboard API puede no estar disponible (permiso denegado, contexto
      // no seguro); la persona siempre puede copiar el texto a mano.
      setCopied(false);
    }
  }

  return (
    <main className="min-h-screen flex items-center justify-center p-4" style={{ backgroundColor: 'var(--color-surface)' }}>
      <div className="w-full max-w-lg overflow-hidden rounded-2xl shadow-xl" style={{ backgroundColor: 'var(--color-card)', color: 'var(--color-text-card)' }}>
        <div aria-hidden="true" className="h-1.5 w-full" style={{ background: 'linear-gradient(90deg, var(--color-primary), var(--color-secondary))' }} />

        <div className="space-y-6 p-8">
          <div className="flex justify-end">
            <SettingsEntry />
          </div>

          <header className="flex flex-col items-center gap-3 text-center">
            <UsbiEmblem />
            <div className="space-y-0.5">
              <h1 className="text-2xl font-bold" style={{ color: 'var(--color-primary)' }}>Crear cuenta</h1>
              <p className="text-sm" style={{ color: 'var(--color-muted)' }}>Universidad Veracruzana</p>
            </div>
          </header>

          <StepIndicator step={step} />

          {error && (
            <p
              role="alert"
              aria-live="assertive"
              className="flex items-start gap-2 rounded-lg border p-3 text-sm"
              style={{ borderColor: 'var(--color-error)', color: 'var(--color-error)', backgroundColor: 'color-mix(in srgb, var(--color-error) 8%, transparent)' }}
            >
              <span aria-hidden="true">⚠</span>
              <span>{error}</span>
            </p>
          )}

          {step === 1 && (
            <>
              {loadingQuestions ? (
                <p className="text-sm text-[--color-muted]" aria-live="polite">Cargando preguntas…</p>
              ) : (
                <form onSubmit={handleAnswersSubmit} className="space-y-4" aria-label="Cuestionario de registro" noValidate>
                  <p className="text-sm text-[--color-muted]">
                    No pedimos nombre, correo ni teléfono. Con estas respuestas generamos tu nickname y contraseña —
                    nadie más las verá.
                  </p>

                  {questions.map((q) => (
                    <Input
                      key={q.id}
                      id={`answer-${q.id}`}
                      label={q.text}
                      value={answers[q.id] ?? ''}
                      onChange={(e) => {
                        // Capturar el valor ANTES de pasarlo a la forma
                        // "updater" de setState: StrictMode invoca ese
                        // updater dos veces para detectar impurezas, y para
                        // la segunda invocación React ya desasoció
                        // e.currentTarget del evento sintético original.
                        const value = e.currentTarget.value;
                        setAnswers((prev) => ({ ...prev, [q.id]: value }));
                      }}
                      error={fieldErrors[q.id]}
                      required
                      aria-required="true"
                      maxLength={200}
                    />
                  ))}

                  <label className="flex items-start gap-2 text-sm">
                    <input
                      type="checkbox"
                      checked={isAdult}
                      onChange={(e) => setIsAdult(e.currentTarget.checked)}
                      className="mt-1 h-5 w-5 shrink-0"
                    />
                    <span>Soy mayor de edad (18 años o más). Puedes actualizar esto después.</span>
                  </label>

                  {loadingPrivacyNotice && (
                    <p className="text-sm text-[--color-muted]" aria-live="polite">Cargando aviso de privacidad…</p>
                  )}
                  {privacyNoticeError && (
                    <p className="text-sm text-[--color-error]" role="alert">{privacyNoticeError}</p>
                  )}
                  {privacyNotice && <PrivacyNoticeInline notice={privacyNotice} />}

                  <label className="flex items-start gap-2 text-sm">
                    <input
                      type="checkbox"
                      checked={acceptedPrivacy}
                      onChange={(e) => setAcceptedPrivacy(e.currentTarget.checked)}
                      required
                      aria-required="true"
                      className="mt-1 h-5 w-5 shrink-0"
                    />
                    <span>He leído y acepto el aviso de privacidad de arriba.</span>
                  </label>

                  <Button
                    type="submit"
                    size="lg"
                    className="w-full font-bold"
                    disabled={loading || questions.length === 0 || !privacyNotice}
                    aria-busy={loading}
                  >
                    {loading && <Spinner />}
                    {loading ? 'Validando…' : 'Continuar'}
                  </Button>
                </form>
              )}
            </>
          )}

          {step === 2 && (
            <form onSubmit={handleConfirmSubmit} className="space-y-4" aria-label="Elegir nickname" noValidate>
              <p className="text-sm text-[--color-muted]">
                Elige uno de estos nicknames generados a partir de tus respuestas. Será tu usuario para iniciar
                sesión.
              </p>

              <fieldset className="space-y-2">
                <legend className="sr-only">Candidatos de nickname</legend>
                {nicknameCandidates.map((candidate) => (
                  <label
                    key={candidate}
                    className="flex items-center gap-3 rounded-lg border p-3 cursor-pointer"
                    style={{
                      borderColor: chosenNickname === candidate ? 'var(--color-primary)' : 'var(--color-border)',
                      backgroundColor: chosenNickname === candidate ? 'color-mix(in srgb, var(--color-primary) 8%, transparent)' : 'transparent',
                    }}
                  >
                    <input
                      type="radio"
                      name="chosen-nickname"
                      value={candidate}
                      checked={chosenNickname === candidate}
                      onChange={() => setChosenNickname(candidate)}
                      className="h-5 w-5 shrink-0"
                    />
                    <span className="font-mono text-base">{candidate}</span>
                  </label>
                ))}
              </fieldset>

              <div className="flex gap-2">
                <Button type="button" variant="outline" onClick={() => setStep(1)} disabled={loading}>
                  Volver
                </Button>
                <Button type="submit" size="lg" className="flex-1 font-bold" disabled={loading || !chosenNickname} aria-busy={loading}>
                  {loading && <Spinner />}
                  {loading ? 'Creando cuenta…' : 'Crear cuenta'}
                </Button>
              </div>
            </form>
          )}

          {step === 3 && (
            <div className="space-y-4">
              <p
                className="rounded-lg border p-3 text-sm"
                style={{ borderColor: 'var(--color-warning)', color: 'var(--color-warning)', backgroundColor: 'color-mix(in srgb, var(--color-warning) 8%, transparent)' }}
              >
                Guarda estos datos ahora — no volverán a mostrarse.
              </p>

              <dl className="space-y-3 rounded-lg border p-4" style={{ borderColor: 'var(--color-border)' }}>
                <div>
                  <dt className="text-xs uppercase tracking-wide text-[--color-muted]">Nickname</dt>
                  <dd className="font-mono text-lg">{issuedNickname}</dd>
                </div>
                <div>
                  <dt className="text-xs uppercase tracking-wide text-[--color-muted]">Contraseña</dt>
                  <dd className="font-mono text-lg">{issuedPassword}</dd>
                </div>
                <div>
                  <dt className="text-xs uppercase tracking-wide text-[--color-muted]">Tu alias en el juego</dt>
                  <dd className="text-lg">{displayAlias}</dd>
                </div>
              </dl>

              <Button type="button" variant="outline" className="w-full" onClick={() => void handleCopyCredentials()}>
                {copied ? 'Copiado ✓' : 'Copiar nickname y contraseña'}
              </Button>

              <Button
                type="button"
                size="lg"
                className="w-full font-bold"
                onClick={() => navigate('/login', { state: { message: 'Cuenta creada. Inicia sesión con tu nuevo nickname.' } })}
              >
                Ir a iniciar sesión
              </Button>
            </div>
          )}

          {step === 1 && (
            <div className="text-center text-sm">
              <Link to="/login" style={{ color: 'var(--color-primary)' }} className="inline-flex items-center justify-center py-2 hover:underline">
                ¿Ya tienes cuenta? Inicia sesión
              </Link>
            </div>
          )}
        </div>
      </div>
    </main>
  );
}

function StepIndicator({ step }: { step: Step }) {
  const labels = ['Cuestionario', 'Nickname', 'Listo'];
  return (
    <ol className="flex items-center justify-center gap-2 text-xs" aria-label="Progreso del registro">
      {labels.map((label, idx) => {
        const n = (idx + 1) as Step;
        const active = n === step;
        const done = n < step;
        return (
          <li key={label} className="flex items-center gap-2">
            <span
              aria-current={active ? 'step' : undefined}
              className="flex h-6 w-6 items-center justify-center rounded-full font-bold"
              style={{
                backgroundColor: active || done ? 'var(--color-primary)' : 'var(--color-border)',
                color: active || done ? 'var(--color-primary-foreground)' : 'var(--color-muted)',
              }}
            >
              {done ? '✓' : n}
            </span>
            <span style={{ color: active ? 'var(--color-primary)' : 'var(--color-muted)' }}>{label}</span>
            {idx < labels.length - 1 && <span aria-hidden="true" className="text-[--color-muted]">—</span>}
          </li>
        );
      })}
    </ol>
  );
}
