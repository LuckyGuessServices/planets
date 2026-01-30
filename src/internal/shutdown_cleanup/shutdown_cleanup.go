package shutdown_cleanup

import (
	"fmt"
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
	mu.Lock()
	defer mu.Unlock()

	isExecuted = true
	if len(cleanupFnSlice) < 1 {
		return
	}

	funcIDByPositionMap := make(map[int]string, len(cleanupFnIDMap))
	for funcID, position := range cleanupFnIDMap {
		funcIDByPositionMap[position] = funcID
	}

	for i := len(cleanupFnSlice) - 1; i >= 0; i-- {
		// Skip an element, if [Unregister] has been called for that element.
		if nil == cleanupFnSlice[i] {
			continue
		}

		// It's crucial for stack of functions to be processed without panics.
		// If an odd thing happen, let's not panic and debug the issue some time later.
		funcId, isFuncIdFoundInMap := funcIDByPositionMap[i]
		if !isFuncIdFoundInMap {
			funcId = fmt.Sprintf("UNKNOWN FUNCTION #%d", i)
		}

		luglog.Printf("[Shutdown #%d] %s", i, funcId)
		cleanupFnSlice[i]()
	}
}

// Register adds a function to a list fo functions called during application shutdown.
// Call [ExecuteStack] inside the "main()" defer function to invoke all registered functions.
//
// Each function added to the list is decorated with a deferred panic recovery. That means your custom functions may
// panic, but that will not prevent execution of other registered functions.
func Register(funcId string, funcHandler func()) {
	mu.Lock()
	defer mu.Unlock()

	if isExecuted {
		panic("Unable to add a shutdown cleanup function: the cleanup itself is executed already!")
	}

	if _, ok := cleanupFnIDMap[funcId]; ok {
		return
	}

	funcIndex := len(cleanupFnSlice)
	decoratedFuncHandler := func() {
		// Do not let the application panic.
		// Warn about an issue, but then let other registered functions to be executed.
		defer func() {
			if panicVal := recover(); panicVal != nil {
				luglog.Printf("[Shutdown #%d] '%s' function panicked: %#v", funcIndex, funcId, panicVal)
			}
		}()

		funcHandler()
	}
	cleanupFnSlice = append(cleanupFnSlice, decoratedFuncHandler)
	cleanupFnIDMap[funcId] = funcIndex
}

func IsRegistered(funcId string) bool {
	mu.Lock()
	defer mu.Unlock()

	_, isRegistered := cleanupFnIDMap[funcId]

	return isRegistered
}

// Unregister removes function from the register list and returns it (so you could call it manually if needed).
// Failure in retrieving a function handler in any way generates an error
// (for instance, if a function was not registered).
//
// The register list is NOT shortened by this operation. Newly registered functions will occupy new register indices,
// which will be reflected in the shutdown log like that:
//
//	Register("shutdown A", func() { /* Do A */ })
//	Register("shutdown B", func() { /* Do B */ })
//	Unregister("shutdown B")
//	Register("shutdown C", func() { /* Do C */ })
//	ExecuteStack()
//
// The code above generates logs:
//
//	... [Shutdown #3] shutdown C
//	... [Shutdown #1] shutdown A
//
// Note the absence of "Shutdown #2" as its corresponding function slot was nullified during unregistering.
func Unregister(funcId string) (func(), error) {
	mu.Lock()
	defer mu.Unlock()

	funcSliceIndex, isIndexFound := cleanupFnIDMap[funcId]
	if !isIndexFound {
		return nil, fmt.Errorf("'%s' function was not registered", funcId)
	}

	if funcSliceIndex < 0 || funcSliceIndex >= len(cleanupFnSlice) {
		return nil, fmt.Errorf("failed to retrieve '%s' function by slice index #%d", funcId, funcSliceIndex)
	}

	registeredFunc := cleanupFnSlice[funcSliceIndex]
	delete(cleanupFnIDMap, funcId)
	cleanupFnSlice[funcSliceIndex] = nil

	return registeredFunc, nil
}
