---
applyTo: '**/*.sql,**/migrations/**,**/*schema*,**/*model*'
description: 'Core relational database design principles (normalization, keys, indexing, denormalization tradeoffs, naming) to apply when designing or reviewing a schema, migration, or data model.'
---

# Database Design Principles

The golden rule: **store each fact once, and reference it by key.** Every principle
below is a way of protecting that rule, or of deliberately trading it away for
measured performance — never on a hunch.

Sources (retrieved 2026-09-27): [Exasol — Database Design Principles & Best Practices](https://www.exasol.com/hub/database/design-principles/),
[dbSchema — Database Design Best Practices](https://dbschema.com/blog/design/database-design-best-practices-2025/),
[DataCamp — Normalization in SQL (1NF-5NF)](https://www.datacamp.com/tutorial/normalization-in-sql),
[SolarWinds — Normalize vs. Denormalize](https://www.solarwinds.com/database-optimization/normalize-vs-denormalize-database).

## Normalize first, to 3NF

- **1NF** — every column holds one atomic value; no repeating groups or
  comma-packed lists in a single field. — A `phone_numbers: "555-1,555-2"` column
  forces string-splitting in every query that needs a single number; a separate
  `phone_numbers` table with one row per number doesn't.
- **2NF** — every non-key column depends on the *whole* primary key, not part of
  it. — In a table keyed on `(order_id, product_id)`, a `customer_name` column
  depends only on `order_id`; it belongs on the `orders` table, not repeated on every
  line item.
- **3NF** — every non-key column depends *only* on the key, not on another
  non-key column (no transitive dependency). — A `products` table with both
  `category_id` and `category_name` lets the two drift out of sync the moment a
  category is renamed in one row but not the other; keep `category_name` only on
  `categories`, referenced by `category_id`.

3NF is the default target — it removes most update/insert/delete anomalies without
over-fragmenting the schema. Going further (4NF/5NF) is rarely worth the join cost
outside specific multi-valued-dependency cases.

```sql
-- BAD: violates 2NF/3NF — customer_name repeated per line item, category_name
-- duplicated and free to drift from categories.name
CREATE TABLE order_items (
  order_id INT, product_id INT, customer_name TEXT,
  category_id INT, category_name TEXT, qty INT
);

-- GOOD: each fact lives in exactly one place, referenced by key
CREATE TABLE orders (order_id INT PRIMARY KEY, customer_id INT REFERENCES customers);
CREATE TABLE categories (category_id INT PRIMARY KEY, name TEXT);
CREATE TABLE products (product_id INT PRIMARY KEY, category_id INT REFERENCES categories);
CREATE TABLE order_items (
  order_id INT REFERENCES orders,
  product_id INT REFERENCES products,
  qty INT,
  PRIMARY KEY (order_id, product_id)
);
```

## Keys and constraints are not optional

- Give every table a **primary key** and every relationship a **foreign key**. —
  Without a foreign key, nothing stops an `order_items.product_id` from pointing at a
  deleted product; the constraint is what makes that impossible instead of merely
  unlikely.
- Enforce invariants the database can check (`NOT NULL`, `UNIQUE`, `CHECK`) rather
  than only in application code. — Application-only validation is bypassed by a
  second service, a script, or a direct psql session; a constraint holds regardless
  of which code path writes the row.
- Prefer a surrogate key (`id SERIAL`/`UUID`) over a natural key that might need to
  change. — A schema keyed on `email` breaks every foreign reference the day a user
  changes their email; an immutable surrogate key doesn't.

## Index for the queries you actually run

- Index columns used in `WHERE`, `JOIN`, and `ORDER BY` — not every column, and not
  by symmetry with the schema. — An index on every column looks thorough but slows
  every write for indexes that no query ever uses; each unused index is pure cost.
- Order composite indexes to match the predicate order of the query that needs
  them. — An index on `(status, created_at)` doesn't efficiently serve a query
  filtering on `created_at` alone; column order in a composite index must match how
  it will actually be queried.
- Remember indexes speed up reads and slow down writes — verify with `EXPLAIN`/
  `EXPLAIN ANALYZE` that a new index is actually used before keeping it, rather than
  adding it and assuming.

## Denormalize only after measuring, never on a hunch

- Stay normalized until `EXPLAIN ANALYZE` proves a specific join is the measured
  bottleneck for a real query. — Premature denormalization is the most common
  schema-design mistake: it doubles write amplification (every write now updates two
  places) and opens the door to the exact data-integrity bugs normalization exists
  to prevent, in exchange for a performance problem that may not even exist yet.
- When denormalizing, denormalize the smallest thing that fixes the measured query
  (one derived column, or a materialized view) rather than flattening the whole
  schema. — A materialized view refreshed on a schedule keeps the source tables
  normalized while still serving the fast read path; a fully denormalized clone of
  the schema doesn't.

## Naming and evolution

- Pick one naming convention (commonly `snake_case`: `user_id`, `created_at`,
  `order_status`) and apply it across the entire schema, no exceptions. — Mixed
  `userId`/`user_id`/`UserID` across tables forces every query author to guess which
  form a given table uses.
- Ship schema changes as versioned, reversible migrations, never as a hand-run
  `ALTER TABLE` against production. — An unrecorded manual change means the schema in
  production and the schema in migration history silently disagree, and nothing can
  reliably reconstruct or roll back the change later.
- Treat a column rename/type change as backward-incompatible: add the new column,
  backfill, migrate readers, then drop the old one — don't rename in place under
  live traffic. — An in-place rename breaks every in-flight query and every
  not-yet-deployed service still referencing the old name, simultaneously.

## Applying these under real constraints

- Model the actual domain and its real query patterns, not a textbook-generic
  shape — a principle applied without reference to how the data is actually read and
  written produces a schema that's "correct" and still wrong for the workload.
- When reviewing a schema or migration, name which principle is at stake and the
  concrete failure it prevents (as in the examples above), not just "this isn't
  normalized" or "this needs an index."
