package shutdown_cleanup

import "sync"

var (
	mu             sync.Mutex
	isExecuted     = false
	cleanupFnSlice []func()
	cleanupFnIdMap = make(map[string]int)
)

// ExecuteStack executes registered functions in LIFO order - starting from the function registered last.
func ExecuteStack() {
	isExecuted = true

	for i := len(cleanupFnSlice) - 1; i >= 0; i-- {
		cleanupFnSlice[i]()
	}
}

func Register(funcId string, f func()) {
	if isExecuted {
		panic("Unable to add a shutdown cleanup function: the cleanup itself is executed already!")
	}
	mu.Lock()
	defer mu.Unlock()

	if _, ok := cleanupFnIdMap[funcId]; ok {
		return
	}

	cleanupFnSlice = append(cleanupFnSlice, f)
	cleanupFnIdMap[funcId] = len(cleanupFnSlice) - 1
}
