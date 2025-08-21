package queue

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestPriorityQueue(t *testing.T) {
	t.Run("NewPriorityQueue", func(t *testing.T) {
		pq := NewPriorityQueue(10)
		assert.NotNil(t, pq)
		assert.Equal(t, 0, pq.Len())
		assert.Equal(t, 10, pq.capacity)
	})

	t.Run("Update and GetTop", func(t *testing.T) {
		pq := NewPriorityQueue(3)

		// Add products
		id1 := uuid.New()
		id2 := uuid.New()
		id3 := uuid.New()
		id4 := uuid.New()

		pq.Update(id1, 100)
		pq.Update(id2, 200)
		pq.Update(id3, 50)

		// Check top products
		top := pq.GetTop()
		assert.Len(t, top, 3)
		assert.Equal(t, int64(200), top[0].ViewCount)
		assert.Equal(t, int64(100), top[1].ViewCount)
		assert.Equal(t, int64(50), top[2].ViewCount)

		// Try to add a product with lower view count (should not be added)
		pq.Update(id4, 25)
		top = pq.GetTop()
		assert.Len(t, top, 3)
		assert.Equal(t, int64(200), top[0].ViewCount)

		// Add a product with higher view count (should replace minimum)
		id5 := uuid.New()
		pq.Update(id5, 150)
		top = pq.GetTop()
		assert.Len(t, top, 3)
		assert.Equal(t, int64(200), top[0].ViewCount)
		assert.Equal(t, int64(150), top[1].ViewCount)
		assert.Equal(t, int64(100), top[2].ViewCount)
	})

	t.Run("Update existing product", func(t *testing.T) {
		pq := NewPriorityQueue(3)

		id1 := uuid.New()
		id2 := uuid.New()

		pq.Update(id1, 100)
		pq.Update(id2, 200)

		// Update existing product
		pq.Update(id1, 300)

		top := pq.GetTop()
		assert.Len(t, top, 2)
		assert.Equal(t, int64(300), top[0].ViewCount)
		assert.Equal(t, id1, top[0].ProductID)
	})

	t.Run("Clear", func(t *testing.T) {
		pq := NewPriorityQueue(5)

		pq.Update(uuid.New(), 100)
		pq.Update(uuid.New(), 200)
		assert.Equal(t, 2, pq.Len())

		pq.Clear()
		assert.Equal(t, 0, pq.Len())
		assert.Empty(t, pq.GetTop())
	})

	t.Run("Capacity enforcement", func(t *testing.T) {
		// Test max capacity of 100
		pq := NewPriorityQueue(200)
		assert.Equal(t, 100, pq.capacity)

		// Test negative capacity defaults to 100
		pq = NewPriorityQueue(-1)
		assert.Equal(t, 100, pq.capacity)
	})
}
