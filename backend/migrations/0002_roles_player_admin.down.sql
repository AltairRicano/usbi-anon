-- No puede revertir el UPDATE de la migración up (qué cuentas eran
-- 'operator'/'director' antes de degradarse a 'admin' no queda registrado en
-- ningún lado): esto solo repone el CHECK de cuatro valores para que el
-- esquema vuelva a aceptarlos, no restaura el rol original de ninguna cuenta.
ALTER TABLE accounts DROP CONSTRAINT accounts_role_check;
ALTER TABLE accounts ADD CONSTRAINT accounts_role_check
    CHECK (role IN ('player', 'admin', 'operator', 'director'));
