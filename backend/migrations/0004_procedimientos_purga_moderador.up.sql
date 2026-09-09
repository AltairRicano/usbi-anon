-- F3 (pools jugador/moderador): CancelAccount (autoservicio de jugador,
-- DELETE /auth/me) necesita borrar account_quiz_answers y anular
-- actor_account_id en audit_log, pero 00_roles_unificado.sql deliberadamente
-- no le da a usbi_app ni DELETE en account_quiz_answers ni UPDATE en
-- audit_log — ver comentarios de esas dos REVOKE en ese archivo.
--
-- En vez de ampliar esos GRANT (lo que le daría a usbi_app la capacidad
-- general de tocar esas columnas, no solo durante una cancelación), estas dos
-- operaciones se aíslan en funciones SECURITY DEFINER propiedad de
-- usbi_moderador: usbi_app solo puede EJECUTARLAS, nunca hacer el DELETE/
-- UPDATE por su cuenta. Al ser funciones SQL corridas dentro de la misma
-- transacción de CancelAccount, no hace falta una segunda conexión ni un
-- segundo pool — siguen viviendo en la Tx única de usbi_app; la función
-- corre server-side con los privilegios de su dueño.

CREATE FUNCTION purge_account_quiz_answers(p_account_id uuid)
RETURNS void
LANGUAGE sql
SECURITY DEFINER
SET search_path = public
AS $$
    DELETE FROM account_quiz_answers WHERE user_id = p_account_id;
$$;

CREATE FUNCTION null_user_in_pseudonymizable_ledgers(p_account_id uuid)
RETURNS void
LANGUAGE sql
SECURITY DEFINER
SET search_path = public
AS $$
    UPDATE experience_history SET user_id = NULL WHERE user_id = p_account_id;
    UPDATE audit_log SET actor_account_id = NULL WHERE actor_account_id = p_account_id;
$$;

ALTER FUNCTION purge_account_quiz_answers(uuid) OWNER TO usbi_moderador;
ALTER FUNCTION null_user_in_pseudonymizable_ledgers(uuid) OWNER TO usbi_moderador;

GRANT EXECUTE ON FUNCTION purge_account_quiz_answers(uuid) TO usbi_app;
GRANT EXECUTE ON FUNCTION null_user_in_pseudonymizable_ledgers(uuid) TO usbi_app;
