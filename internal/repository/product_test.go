package repository_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
	"github.com/stretchr/testify/assert"
	"github.com/tushar-kalsi/product-views/internal/repository"
)

var dbPool *pgxpool.Pool
var sqlDB *sql.DB

func TestMain(m *testing.M) {
	// Uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		panic(fmt.Sprintf("Could not connect to docker: %s", err))
	}

	// Pull the Postgres image, create a container with the necessary configurations
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "14-alpine",
		Env: []string{
			"POSTGRES_USER=test",
			"POSTGRES_PASSWORD=test",
			"POSTGRES_DB=testdb",
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		// Set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		panic(fmt.Sprintf("Could not start resource: %s", err))
	}

	// Exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 120 * time.Second
	if err = pool.Retry(func() error {
		var err error
		connString := fmt.Sprintf("postgres://test:test@localhost:%s/testdb?sslmode=disable", resource.GetPort("5432/tcp"))

		// Create pgxpool connection for migrations
		dbPool, err = pgxpool.New(context.Background(), connString)
		if err != nil {
			return err
		}

		// Create standard sql.DB connection for the repository
		sqlDB = stdlib.OpenDB(*dbPool.Config().ConnConfig)

		return dbPool.Ping(context.Background())
	}); err != nil {
		panic(fmt.Sprintf("Could not connect to docker: %s", err))
	}

	// Run migrations using pgxpool
	_, err = dbPool.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS products (
            id UUID PRIMARY KEY,
            name VARCHAR(255) NOT NULL,
            description TEXT,
            view_count BIGINT NOT NULL DEFAULT 0,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        )`)
	if err != nil {
		panic(fmt.Sprintf("Failed to create test table: %v", err))
	}

	// Run the tests
	code := m.Run()

	// Clean up
	sqlDB.Close()
	dbPool.Close()
	if err := pool.Purge(resource); err != nil {
		panic(fmt.Sprintf("Could not purge resource: %s", err))
	}

	os.Exit(code)
}

func TestProductRepository(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewProductRepository(sqlDB)

	// Create a test product
	product := &repository.Product{
		Name:        "Test Product",
		Description: "Test Description",
		ViewCount:   0,
	}

	t.Run("CreateProduct", func(t *testing.T) {
		err := repo.CreateProduct(ctx, product)
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, product.ID)
	})

	t.Run("GetProduct", func(t *testing.T) {
		retrieved, err := repo.GetProduct(ctx, product.ID)
		assert.NoError(t, err)
		assert.Equal(t, product.ID, retrieved.ID)
		assert.Equal(t, product.Name, retrieved.Name)
		assert.Equal(t, product.Description, retrieved.Description)
		assert.Equal(t, product.ViewCount, retrieved.ViewCount)
	})

	t.Run("IncrementViewCount", func(t *testing.T) {
		err := repo.IncrementViewCount(ctx, product.ID)
		assert.NoError(t, err)

		updated, err := repo.GetProduct(ctx, product.ID)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), updated.ViewCount)
	})

	t.Run("GetTopViewedProducts", func(t *testing.T) {
		// Create a few more test products
		products := []*repository.Product{
			{Name: "Product 1", ViewCount: 5},
			{Name: "Product 2", ViewCount: 10},
			{Name: "Product 3", ViewCount: 3},
		}

		for _, p := range products {
			p := p // Create a new variable for the loop
			err := repo.CreateProduct(ctx, p)
			assert.NoError(t, err)

			// Increment view count to set the desired view count
			for i := int64(0); i < p.ViewCount; i++ {
				err := repo.IncrementViewCount(ctx, p.ID)
				assert.NoError(t, err)
			}
		}

		// Test getting top 2 products
		topProducts, err := repo.GetTopViewedProducts(ctx, 2)
		assert.NoError(t, err)
		assert.Len(t, topProducts, 2)
		assert.Equal(t, "Product 2", topProducts[0].Name) // Should be first because it has the most views
		assert.Equal(t, "Product 1", topProducts[1].Name)

		// Test getting more products than exist
		allProducts, err := repo.GetTopViewedProducts(ctx, 100)
		assert.NoError(t, err)
		assert.True(t, len(allProducts) >= 4) // At least 4 products now
	})

	t.Run("NonExistentProduct", func(t *testing.T) {
		nonExistentID := uuid.New()

		// Test GetProduct with non-existent ID
		_, err := repo.GetProduct(ctx, nonExistentID)
		assert.Error(t, err)
		assert.Equal(t, "product not found", err.Error())

		// Test IncrementViewCount with non-existent ID
		err = repo.IncrementViewCount(ctx, nonExistentID)
		assert.Error(t, err)
	})

	t.Run("EmptyTopProducts", func(t *testing.T) {
		// Test with limit 0 (should return empty slice)
		products, err := repo.GetTopViewedProducts(ctx, 0)
		assert.NoError(t, err)
		assert.Empty(t, products)
	})
}
