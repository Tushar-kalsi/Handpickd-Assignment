package repository

import (
    "context"
    "database/sql"
    "errors"
    "github.com/google/uuid"
    "time"
)

// Product represents a product in the database
type Product struct {
    ID          uuid.UUID `db:"id"`
    Name        string    `db:"name"`
    Description string    `db:"description"`
    ViewCount   int64     `db:"view_count"`
    CreatedAt   time.Time `db:"created_at"`
    UpdatedAt   time.Time `db:"updated_at"`
}

// ProductRepository defines the interface for product data operations
type ProductRepository interface {
    IncrementViewCount(ctx context.Context, productID uuid.UUID) error
    GetTopViewedProducts(ctx context.Context, limit int) ([]Product, error)
    GetProduct(ctx context.Context, id uuid.UUID) (*Product, error)
    CreateProduct(ctx context.Context, p *Product) error
}

type productRepository struct {
    db *sql.DB
}

// NewProductRepository creates a new ProductRepository
func NewProductRepository(db *sql.DB) ProductRepository {
    return &productRepository{db: db}
}

// IncrementViewCount increments the view count for a product
func (r *productRepository) IncrementViewCount(ctx context.Context, productID uuid.UUID) error {
    query := `
        UPDATE products 
        SET view_count = view_count + 1 
        WHERE id = $1`

    result, err := r.db.ExecContext(ctx, query, productID)
    if err != nil {
        return err
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }

    if rowsAffected == 0 {
        return errors.New("product not found")
    }

    return nil
}

// GetTopViewedProducts returns the top N most viewed products
func (r *productRepository) GetTopViewedProducts(ctx context.Context, limit int) ([]Product, error) {
    if limit > 100 {
        limit = 100 // Enforce max limit
    }

    query := `
        SELECT id, name, description, view_count, created_at, updated_at
        FROM products
        ORDER BY view_count DESC
        LIMIT $1`

    rows, err := r.db.QueryContext(ctx, query, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var products []Product
    for rows.Next() {
        var p Product
        err := rows.Scan(
            &p.ID,
            &p.Name,
            &p.Description,
            &p.ViewCount,
            &p.CreatedAt,
            &p.UpdatedAt,
        )
        if err != nil {
            return nil, err
        }
        products = append(products, p)
    }

    if err = rows.Err(); err != nil {
        return nil, err
    }

    return products, nil
}

// GetProduct retrieves a product by ID
func (r *productRepository) GetProduct(ctx context.Context, id uuid.UUID) (*Product, error) {
    query := `
        SELECT id, name, description, view_count, created_at, updated_at
        FROM products
        WHERE id = $1`

    var p Product
    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &p.ID,
        &p.Name,
        &p.Description,
        &p.ViewCount,
        &p.CreatedAt,
        &p.UpdatedAt,
    )

    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, errors.New("product not found")
        }
        return nil, err
    }

    return &p, nil
}

// CreateProduct creates a new product
func (r *productRepository) CreateProduct(ctx context.Context, p *Product) error {
    query := `
        INSERT INTO products (name, description, view_count)
        VALUES ($1, $2, $3)
        RETURNING id, created_at, updated_at`

    return r.db.QueryRowContext(
        ctx,
        query,
        p.Name,
        p.Description,
        p.ViewCount,
    ).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}
