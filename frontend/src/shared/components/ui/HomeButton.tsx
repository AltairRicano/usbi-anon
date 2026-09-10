import { LinkButton } from './LinkButton';

// Botón estandarizado de "volver al menú principal". Antes cada pantalla
// tenía su propia etiqueta ("Inicio"/"Dashboard"/"Volver") y anidaba un
// <Link> dentro de <Button> (=> <a> dentro de <button>, HTML inválido, el
// navegador lo repara duplicando el elemento interactivo). Este componente
// fija texto y destino en un solo lugar para toda la app.
export function HomeButton({ className }: { className?: string }) {
  return (
    <LinkButton to="/" variant="outline" size="sm" className={className}>
      Inicio
    </LinkButton>
  );
}
