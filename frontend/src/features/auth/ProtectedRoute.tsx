import { type JSX } from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { useAuthStore } from './useAuthStore';
import type { z } from 'zod';
import type { UserRoleSchema } from '../../shared/schemas';

type UserRole = z.infer<typeof UserRoleSchema>;

interface ProtectedRouteProps {
  children: JSX.Element;
  allowedRoles?: UserRole[];
}

/**
 * Protege una ruta verificando autenticación y, opcionalmente, el rol.
 * Redirige a /login preservando la ruta de origen para volver tras el login.
 */
export function ProtectedRoute({ children, allowedRoles }: ProtectedRouteProps): JSX.Element {
  const { isAuthenticated, user } = useAuthStore();
  const location = useLocation();

  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  if (allowedRoles && user && !allowedRoles.includes(user.role)) {
    return <Navigate to="/unauthorized" replace />;
  }

  return children;
}
