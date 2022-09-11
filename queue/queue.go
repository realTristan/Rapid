package queue

// Import Packages
import "sync"

//////////////////////////////////////////////////////
// For the official queue system,
// visit https://github.com/realTristan/GoQueue
//////////////////////////////////////////////////////

// Type Item interface{}
//
//	 The 'Item' Type is the type of variables that will be going inside the queue slice
//	 The Item is declared as interface so it is possible to have multiple types
//		   within the Queue Slice
type Item interface{}

// type ItemQueue struct
//
//	 The 'ItemQueue' Struct contains the []'Type Item interface{}' slice
//	 This struct holds two keys,
//	    - items -> the []'Type Item interface{}' slice
//	    - mutex -> the mutex lock which prevents overwrites and data corruption
//				  ↳ We use RWMutex instead of Mutex as it's better for majority read slices
type ItemQueue struct {
	items []Item
	mutex *sync.RWMutex
}

// Create() -> *ItemQueue
// The Create() function will return an empty ItemQueue
func Create() *ItemQueue {
	return &ItemQueue{mutex: &sync.RWMutex{}, items: []Item{}}
}

// q.Remove(Item) -> None
// The Remove() function will secure the ItemQueue before iterating
// through said ItemQueue and remove the given Item (_item)
func (q *ItemQueue) Remove(item Item) {
	// Lock/Unlock the mutex
	q.mutex.Lock()
	defer q.mutex.Unlock()

	// Iterate over the queue
	for i := 0; i < len(q.items); i++ {
		if q.items[i] == item {
			q.items = append(q.items[:i], q.items[i+1:]...)
			return
		}
	}
}

// q.Put(Item) -> None
// The Put() function is used to add a new item to the provided ItemQueue
func (q *ItemQueue) Put(i Item) {
	// Lock/Unlock the mutex
	q.mutex.Lock()
	defer q.mutex.Unlock()

	// Add the item
	q.items = append(q.items, i)
}

// q.Get() -> Item
// The Get() function will append the first item of the ItemQueue to the back of the slice
// then remove it from the front
// The function returns the first item of the ItemQueue
func (q *ItemQueue) Get() Item {
	// Get the item from the queue
	var item Item = q.items[0]

	// Modify the queue
	q.items = append(q.items, q.items[0])
	q.items = q.items[1:]

	// Return the item
	return item
}

// q.Grab() -> Item
// The Grab() function will return the first item of the ItemQueue then
// remove it from said ItemQueue
func (q *ItemQueue) Grab() Item {
	// Lock/Unlock the mutex
	q.mutex.Lock()
	defer q.mutex.Unlock()

	// Grah the item from the queue
	var item Item = q.items[0]
	q.items = q.items[1:]

	// Return the item
	return item
}

// q.IsNotEmpty() -> bool
// The IsNotEmpty() function will return whether the provided ItemQueue contains any Items
func (q *ItemQueue) IsNotEmpty() bool {

	// Lock Reading
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	// Return whether length is greater than 0
	return len(q.items) > 0
}

// q.Size() -> int
// The Size() function will return the length of the ItemQueue slice
func (q *ItemQueue) Size() int {

	// Lock Reading
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	// Return the queue length
	return len(q.items)
}
