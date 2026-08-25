import { Link } from 'react-router-dom';
import { Button } from '../../shared/components/ui/Button';
import { Card, CardContent, CardHeader, CardTitle } from '../../shared/components/ui/Card';
import { useAuthStore } from '../auth/useAuthStore';

/**
 * Landing mínima tras el login. F10 solo cubre auth + paneles de admin
 * (plan/04_Rediseno_identidad_gustos.md §5) — el dashboard de progreso,
 * niveles y minijuegos de ../usbi/frontend se porta en una fase posterior,
 * no listada todavía en el plan de fases F5–F11.
 */
export default function HomePage() {
  const user = useAuthStore((s) => s.user);
  const logout = useAuthStore((s) => s.logout);
  const isAdmin = user?.role === 'admin';

  return (
    <main className="min-h-screen p-6" style={{ backgroundColor: 'var(--color-surface)' }}>
      <div className="mx-auto max-w-2xl space-y-6">
        <header className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold">Hola, {user?.display_alias ?? user?.nickname}</h1>
            <p className="text-sm text-[--color-muted]">Rol: {user?.role}</p>
          </div>
          <Button variant="outline" onClick={logout}>Cerrar sesión</Button>
        </header>

        <Card>
          <CardHeader>
            <CardTitle>Progreso y minijuegos</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-[--color-muted]">
              Esta pantalla es un punto de aterrizaje mínimo para probar el registro y el login de punta a punta.
              El dashboard de progreso, niveles y minijuegos todavía no se ha portado a este proyecto.
            </p>
          </CardContent>
        </Card>

        {isAdmin && (
          <Card>
            <CardHeader>
              <CardTitle>Administración</CardTitle>
            </CardHeader>
            <CardContent className="flex flex-wrap gap-2">
              <Button variant="outline">
                <Link to="/admin/registration-questions">Banco de preguntas</Link>
              </Button>
              <Button variant="outline">
                <Link to="/admin/accounts">Cuentas</Link>
              </Button>
            </CardContent>
          </Card>
        )}
      </div>
    </main>
  );
}
