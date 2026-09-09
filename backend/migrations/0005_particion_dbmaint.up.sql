-- F3 (pools jugador/moderador): internal/dbmaint es el único componente del
-- backend que hace DDL en tiempo de ejecución (CREATE TABLE ... PARTITION OF
-- sobre level_attempts/daily_streak). Ni usbi_app ni usbi_moderador pueden
-- tener privilegios de DDL, y crear un tercer pool con la credencial de
-- migración (dueña del esquema) dentro del mismo binario HTTP sería
-- exactamente el tipo de permiso de sobra que esta separación de roles busca
-- evitar.
--
-- Solución: una función SECURITY DEFINER que hace la única operación DDL que
-- dbmaint necesita, con el nombre de tabla validado contra una lista fija
-- (nunca interpolado libremente) y el año acotado a un rango razonable. Un
-- rol nuevo, usbi_dbmaint (00_roles_unificado.sql), solo puede EXECUTE esta
-- función — cero GRANT de tabla. Mismo patrón que
-- purge_account_quiz_answers/null_user_in_pseudonymizable_ledgers
-- (migración 0004): la función corre con los privilegios de quien la creó
-- (quien aplicó las migraciones), no con los del rol que la invoca.

CREATE FUNCTION ensure_yearly_partition(p_table text, p_year integer)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
BEGIN
    IF p_table NOT IN ('level_attempts', 'daily_streak') THEN
        RAISE EXCEPTION 'ensure_yearly_partition: tabla no permitida: %', p_table;
    END IF;
    IF p_year < 2000 OR p_year > 2100 THEN
        RAISE EXCEPTION 'ensure_yearly_partition: año fuera de rango: %', p_year;
    END IF;

    EXECUTE format(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF %I FOR VALUES FROM (%L) TO (%L)',
        p_table || '_' || p_year, p_table,
        p_year || '-01-01', (p_year + 1) || '-01-01'
    );
END;
$$;

GRANT EXECUTE ON FUNCTION ensure_yearly_partition(text, integer) TO usbi_dbmaint;
