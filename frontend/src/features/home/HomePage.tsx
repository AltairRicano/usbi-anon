import { Link } from 'react-router-dom';
import { Button } from '../../shared/components/ui/Button';
import { Card, CardContent, CardHeader, CardTitle } from '../../shared/components/ui/Card';
import { SettingsEntry } from '../../shared/components/SettingsEntry';
import { useAuthStore } from '../auth/useAuthStore';

/**
 * Landing mínima tras el login. El dashboard de progreso, niveles y
 * minijuegos (F10.10, plan/05_Contenido_maker_y_juego.md) todavía no se ha
 * portado — el maker local (F10.8) sí, y por eso ya tiene entrada aquí.
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
          <div className="flex items-center gap-2">
            <SettingsEntry />
            <Button variant="outline" onClick={logout}>Cerrar sesión</Button>
          </div>
        </header>

        <Card>
          <CardHeader>
            <CardTitle>Progreso y minijuegos</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-[--color-muted]">
              Esta pantalla es un punto de aterrizaje mínimo para probar el registro y el login de punta a punta.
              El dashboard de progreso y los niveles jugables todavía no se han portado a este proyecto.
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Maker local</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            <p className="text-sm text-[--color-muted]">
              Crea niveles de prueba y guárdalos en este navegador, o expórtalos como archivo JSON.
            </p>
            <Button variant="outline">
              <Link to="/maker">Abrir maker local</Link>
            </Button>
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
              <Button variant="outline">
                <Link to="/admin/content">Contenido</Link>
              </Button>
            </CardContent>
          </Card>
        )}
      </div>
    </main>
  );
}
