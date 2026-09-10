import { Link, type LinkProps } from 'react-router-dom';
import { buttonClasses, type ButtonSize, type ButtonVariant } from '../../utils';

export interface LinkButtonProps extends LinkProps {
  variant?: ButtonVariant;
  size?: ButtonSize;
}

// Botón que navega con react-router en vez de disparar un onClick. Reemplaza
// el patrón <Button><Link>...</Link></Button> usado antes en toda la app:
// ese patrón anida un <a> dentro de un <button> (HTML inválido) y el
// navegador lo repara moviendo el <a> fuera del <button>, duplicando el
// elemento interactivo en el DOM. LinkButton renderiza un único <a>
// estilizado igual que <Button>.
export function LinkButton({ variant = 'outline', size = 'sm', className, children, ...props }: LinkButtonProps) {
  return (
    <Link className={buttonClasses(variant, size, className)} {...props}>
      {children}
    </Link>
  );
}
