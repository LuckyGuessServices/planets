package orm

// Ptr lets setting a value to a pointer variable in a single line.
//
//	var created_at *time.Time
//	created_at = Ptr[time.Time](time.Now())
//
// Instead of:
//
//	var created_at *time.Time
//	currentTime := time.Now()
//	created_at = &currentTime
func Ptr[Type any](value Type) *Type {
	return &value
}
