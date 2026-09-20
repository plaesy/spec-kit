---
description: 'Guidelines for generating SQL statements and stored procedures'
applyTo: '**/*.sql'
---

# SQL Development

## Database schema generation
- Singular table and column names
- Primary key column named `id` on every table
- `created_at`/`updated_at` timestamp columns on every table

## Database schema design
- Primary key constraint on every table
- Named, inline foreign key constraints, referencing the parent table's primary key, with `ON DELETE CASCADE` and `ON UPDATE CASCADE`

## SQL Coding Style
- Uppercase SQL keywords (SELECT, FROM, WHERE)
- Consistent indentation for nested queries/conditions
- Comments for complex logic
- Break long queries across multiple lines
- Consistent clause order (SELECT, FROM, JOIN, WHERE, GROUP BY, HAVING, ORDER BY)

## SQL Query Structure
- Explicit column names, never `SELECT *`
- Qualify columns with table name/alias across multiple tables
- Prefer joins over subqueries where possible
- `LIMIT`/`TOP` to restrict result sets
- Index frequently queried columns
- Avoid functions on indexed columns in `WHERE` clauses

## Stored Procedure Naming Conventions
- Prefix `usp_`, PascalCase, descriptive of purpose (e.g. `usp_GetCustomerOrders`)
- Plural noun for multi-record returns (`usp_GetProducts`), singular for single-record (`usp_GetProduct`)

## Parameter Handling
- Prefix `@`, camelCase names
- Default values for optional parameters, validated before use
- Document parameters with comments
- Consistent ordering: required first, optional later


## Stored Procedure Structure
- Header comment block: description, parameters, return values
- Standardized error codes/messages
- Consistent result-set column order
- `OUTPUT` parameters for status info
- Temp tables prefixed `tmp_`


## SQL Security Best Practices
- Parameterize all queries (prevent SQL injection); prepared statements for dynamic SQL
- Never embed credentials in SQL scripts
- Proper error handling without exposing system details
- Avoid dynamic SQL inside stored procedures

## Transaction Management
- Explicit begin/commit transactions
- Appropriate isolation levels per requirements
- Avoid long-running, table-locking transactions
- Batch processing for large data operations
- `SET NOCOUNT ON` in data-modifying stored procedures