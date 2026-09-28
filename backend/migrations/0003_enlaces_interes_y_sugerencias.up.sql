-- Migration: 0003_enlaces_interes_y_sugerencias.up.sql
--
-- Sección "Más" / "Enlaces de interés" del frontend: tarjetas deslizables
-- agrupadas por categoría (p. ej. "Encuestas", "Libros relacionados") que
-- redirigen a un link externo, más el buzón de sugerencias anónimo debajo.
--
-- Igual que 0002_roles_player_admin, esta migración vive aparte de
-- 0001_esquema_unificado.up.sql en vez de plegarse ahí: 0001 documenta el
-- baseline ya aplicado contra una base con datos reales (cuentas sembradas,
-- admin), no el estado corriente del esquema — el mismo razonamiento que ya
-- se aplicó para no tocar 0001 con la poda de roles de 0002.

-- ═══════════════════════════════════════════════════════════════════════════
-- interest_link_categories
-- ═══════════════════════════════════════════════════════════════════════════
--
-- Categorías del carrusel, editables por un admin sin tocar código — de ahí
-- que no sea un ENUM ni una constante en Go. display_order controla en qué
-- orden aparece cada carrusel de categoría dentro de la sección; las tarjetas
-- DENTRO de una categoría no tienen orden propio, se listan por created_at
-- (decisión explícita: solo la categoría necesita orden manual).
CREATE TABLE interest_link_categories (
    id            UUID        PRIMARY KEY,
    name          VARCHAR     NOT NULL UNIQUE CHECK (btrim(name) <> ''),
    display_order SMALLINT    NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX interest_link_categories_display_order_idx
    ON interest_link_categories (display_order);

COMMENT ON TABLE interest_link_categories IS
    'Categorías del carrusel de "enlaces de interés" (sección "Más" del '
    'frontend). CRUD completo para un admin (usbi_moderador); un jugador '
    '(usbi_app) solo las lee.';

-- ═══════════════════════════════════════════════════════════════════════════
-- interest_links
-- ═══════════════════════════════════════════════════════════════════════════
--
-- Una tarjeta deslizable del carrusel. ON DELETE RESTRICT en category_id:
-- borrar una categoría con enlaces vivos debe fallar explícitamente, no
-- arrastrarse en cascada — un admin que quiera vaciar una categoría antes de
-- borrarla no debe poder perder enlaces por accidente.
CREATE TABLE interest_links (
    id          UUID         PRIMARY KEY,
    category_id UUID         NOT NULL REFERENCES interest_link_categories(id) ON DELETE RESTRICT,
    title       VARCHAR(50)  NOT NULL CHECK (btrim(title) <> ''),
    description VARCHAR(100) NOT NULL CHECK (btrim(description) <> ''),
    -- Color de la tarjeta en el front, hexadecimal de 6 dígitos con '#'.
    color       VARCHAR(7)   NOT NULL CHECK (color ~ '^#[0-9A-Fa-f]{6}$'),
    -- http(s) explícito: la tarjeta renderiza esto como href — un esquema
    -- distinto (javascript:, data:, …) sembrado por error en el panel admin
    -- no debe ni siquiera poder guardarse.
    url         TEXT         NOT NULL CHECK (url ~* '^https?://'),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX interest_links_category_id_idx ON interest_links (category_id);

COMMENT ON TABLE interest_links IS
    'Tarjetas del carrusel de "enlaces de interés". CRUD completo para un '
    'admin (usbi_moderador); un jugador (usbi_app) solo las lee.';

-- ═══════════════════════════════════════════════════════════════════════════
-- suggestions
-- ═══════════════════════════════════════════════════════════════════════════
--
-- Buzón de sugerencias, deliberadamente SIN vínculo a accounts: es anónimo
-- por diseño, no hay account_id ni ninguna otra FK. levels_completed_snapshot
-- y xp_snapshot son una COPIA congelada en el momento del envío — la misma
-- fórmula que ya usa el perfil del jugador (GetUserProgressTotals: niveles
-- vivos de player_progress + retirados de account_retired_progress, y
-- SUM(experience_history.xp_gained)) — para poder correlacionar "cuánto ha
-- jugado quien sugirió esto" sin poder identificar jamás a la cuenta que lo
-- mandó, ni siquiera cruzando con otra tabla.
CREATE TABLE suggestions (
    id                         UUID          PRIMARY KEY,
    description                VARCHAR(1000) NOT NULL CHECK (btrim(description) <> ''),
    levels_completed_snapshot  INTEGER       NOT NULL CHECK (levels_completed_snapshot >= 0),
    xp_snapshot                INTEGER       NOT NULL CHECK (xp_snapshot >= 0),
    submitted_at               TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

-- Soporta el panel admin, que lista las sugerencias más recientes primero.
CREATE INDEX suggestions_submitted_at_idx ON suggestions (submitted_at DESC);

COMMENT ON TABLE suggestions IS
    'Buzón de sugerencias anónimo: sin account_id ni ninguna otra columna que '
    'permita vincular una fila a una cuenta. Un jugador (usbi_app) solo '
    'inserta; un admin (usbi_moderador) solo lee y borra, nunca inserta ni '
    'actualiza.';
