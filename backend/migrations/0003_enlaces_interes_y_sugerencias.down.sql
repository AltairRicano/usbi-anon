-- Migration: 0003_enlaces_interes_y_sugerencias.down.sql
-- Orden inverso al up.sql: primero lo que no tiene dependientes, luego lo que
-- otras tablas referencian por FK.

DROP INDEX IF EXISTS suggestions_submitted_at_idx;
DROP TABLE IF EXISTS suggestions;

DROP INDEX IF EXISTS interest_links_category_id_idx;
DROP TABLE IF EXISTS interest_links;

DROP INDEX IF EXISTS interest_link_categories_display_order_idx;
DROP TABLE IF EXISTS interest_link_categories;
