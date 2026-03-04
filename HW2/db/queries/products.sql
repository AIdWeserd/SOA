-- name: CreateProduct :one
INSERT INTO products (name, description, price, stock, category, status)
VALUES (
    sqlc.arg('name'),
    sqlc.narg('description'),
    sqlc.arg('price'),
    sqlc.arg('stock'),
    sqlc.arg('category'),
    sqlc.arg('status')
) RETURNING *;

-- name: GetProductByID :one
SELECT id, name, description, price, stock, category, status, created_at, updated_at
FROM products
WHERE id = sqlc.arg('id')
LIMIT 1;

-- name: GetProducts :many
SELECT id, name, description, price, stock, category, status, created_at, updated_at
FROM products
WHERE
    (sqlc.narg('status')::product_status IS NULL OR status = sqlc.narg('status')) AND
    (sqlc.narg('category')::varchar IS NULL OR category = sqlc.narg('category'))
ORDER BY created_at DESC
LIMIT sqlc.arg('limit')
OFFSET sqlc.arg('offset');

-- name: UpdateProduct :one
UPDATE products SET
    name        = sqlc.arg('name'),
    description = sqlc.narg('description'),
    price       = sqlc.arg('price'),
    stock       = sqlc.arg('stock'),
    category    = sqlc.arg('category'),
    status      = sqlc.arg('status'),
    updated_at  = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: ArchiveProduct :one
UPDATE products SET
    status = 'ARCHIVED',
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: CountProducts :one
SELECT COUNT(*)
FROM products
WHERE
    (sqlc.narg('status')::product_status IS NULL OR status = sqlc.narg('status')) AND
    (sqlc.narg('category')::varchar IS NULL OR category = sqlc.narg('category'));
