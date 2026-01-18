package shutdown_cleanup

import (
	"sync"

	"github.com/LuckyGuessServices/planets/internal/luglog"
)

var (
	mu             sync.Mutex
	isExecuted     = false
	cleanupFnSlice []func()
	cleanupFnIDMap = make(map[string]int)
)

// ExecuteStack executes registered functions in LIFO order - starting from the function registered last.
func ExecuteStack() {
	isExecuted = true

	funcIDByPositionMap := make(map[int]string, len(cleanupFnIDMap))
	for funcID, position := range cleanupFnIDMap {
		funcIDByPositionMap[position] = funcID
	}

	for i := len(cleanupFnSlice) - 1; i >= 0; i-- {
		luglog.Printf("[Shutdown #%d] %s", i, funcIDByPositionMap[i])
		cleanupFnSlice[i]()
	}
}

// Register adds a function to a list fo functions called during application shutdown.
// Call [ExecuteStack] inside the "main()" defer function to invoke all registered functions.
func Register(funcId string, f func()) {
	if isExecuted {
		panic("Unable to add a shutdown cleanup function: the cleanup itself is executed already!")
	}
	mu.Lock()
	defer mu.Unlock()

	if _, ok := cleanupFnIDMap[funcId]; ok {
		return
	}

	cleanupFnSlice = append(cleanupFnSlice, f)
	cleanupFnIDMap[funcId] = len(cleanupFnSlice) - 1
}
