package queue

import (
	"container/heap"
	"sync"

	"github.com/google/uuid"
)

// ProductView represents a product with its view count for the priority queue
type ProductView struct {
	ProductID uuid.UUID
	ViewCount int64
	index     int // index in the heap
}

// PriorityQueue implements a min-heap for top N products
type PriorityQueue struct {
	items    []*ProductView
	itemMap  map[uuid.UUID]*ProductView
	capacity int
	mu       sync.RWMutex
}

// NewPriorityQueue creates a new priority queue with specified capacity
func NewPriorityQueue(capacity int) *PriorityQueue {
	if capacity <= 0 || capacity > 100 {
		capacity = 100 // enforce max limit
	}
	pq := &PriorityQueue{
		items:    make([]*ProductView, 0, capacity),
		itemMap:  make(map[uuid.UUID]*ProductView),
		capacity: capacity,
	}
	heap.Init(pq)
	return pq
}

// Len returns the number of items in the queue
func (pq *PriorityQueue) Len() int { return len(pq.items) }

// Less defines the ordering (min-heap based on view count)
func (pq *PriorityQueue) Less(i, j int) bool {
	return pq.items[i].ViewCount < pq.items[j].ViewCount
}

// Swap swaps two items in the queue
func (pq *PriorityQueue) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
	pq.items[i].index = i
	pq.items[j].index = j
}

// Push adds an item to the queue
func (pq *PriorityQueue) Push(x interface{}) {
	n := len(pq.items)
	item := x.(*ProductView)
	item.index = n
	pq.items = append(pq.items, item)
	pq.itemMap[item.ProductID] = item
}

// Pop removes and returns the minimum item
func (pq *PriorityQueue) Pop() interface{} {
	old := pq.items
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	pq.items = old[0 : n-1]
	delete(pq.itemMap, item.ProductID)
	return item
}

// Update adds or updates a product's view count
func (pq *PriorityQueue) Update(productID uuid.UUID, viewCount int64) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	// Check if product already exists in the queue
	if item, exists := pq.itemMap[productID]; exists {
		// Update existing item
		item.ViewCount = viewCount
		heap.Fix(pq, item.index)
		return
	}

	// If queue is not full, add the new item
	if len(pq.items) < pq.capacity {
		newItem := &ProductView{
			ProductID: productID,
			ViewCount: viewCount,
		}
		heap.Push(pq, newItem)
		return
	}

	// Queue is full, check if new item should replace the minimum
	if viewCount > pq.items[0].ViewCount {
		// Remove the minimum item
		heap.Pop(pq)
		// Add the new item
		newItem := &ProductView{
			ProductID: productID,
			ViewCount: viewCount,
		}
		heap.Push(pq, newItem)
	}
}

// GetTop returns the top N products sorted by view count (descending)
func (pq *PriorityQueue) GetTop() []*ProductView {
	pq.mu.RLock()
	defer pq.mu.RUnlock()

	// Create a copy and sort in descending order
	result := make([]*ProductView, len(pq.items))
	for i, item := range pq.items {
		result[i] = &ProductView{
			ProductID: item.ProductID,
			ViewCount: item.ViewCount,
		}
	}

	// Sort in descending order by view count
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].ViewCount < result[j].ViewCount {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}

// Clear removes all items from the queue
func (pq *PriorityQueue) Clear() {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	pq.items = pq.items[:0]
	pq.itemMap = make(map[uuid.UUID]*ProductView)
	heap.Init(pq)
}
