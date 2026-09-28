-- F10.6: los roles 'operator' y 'director' entraron copiados verbatim de
-- ../usbi en F3 (junto con el resto de internal/domain) y nunca se pidieron:
-- no hay una sola comprobación de autorización en el código que distinga
-- 'operator' o 'director' de 'admin'. Se reducen a los dos roles que
-- realmente existen. Ver plan/05_Contenido_maker_y_juego.md §3.

-- Degradar cualquier cuenta con un rol que va a dejar de ser válido, antes
-- de endurecer el CHECK — así la migración no falla si alguna cuenta de
-- prueba quedó sembrada con 'operator' o 'director'.
UPDATE accounts SET role = 'admin' WHERE role IN ('operator', 'director');

ALTER TABLE accounts DROP CONSTRAINT accounts_role_check;
ALTER TABLE accounts ADD CONSTRAINT accounts_role_check
    CHECK (role IN ('player', 'admin'));
