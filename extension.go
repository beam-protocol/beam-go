package beam

// KindExtension defines the type for custom extension keys used in BEAM entries.
type KindExtension string

const (
	// KindCommentsExtension is the key used for storing comments in the entry's extensions.
	KindCommentsExtension KindExtension = "_comments"
	// KindViewsExtension is the key used for storing view counts in the entry's extensions.
	KindViewsExtension KindExtension = "_views"
	// KindLikesExtension is the key used for storing like counts in the entry's extensions.
	KindLikesExtension KindExtension = "_likes"
	// KindSharesExtension is the key used for storing share counts in the entry's extensions.
	KindSharesExtension KindExtension = "_shares"
	// KindRatingsExtension is the key used for storing ratings in the entry's extensions.
	KindRatingsExtension KindExtension = "_ratings"
)

// ExtensionFields represents a map of custom extension fields for an entry.
// These fields can be used to store additional metadata or custom data
// that is not part of the standard BEAM entry structure. Keys should start with an underscore
// to avoid conflicts with standard fields.
// Example: {"_customField": "value", "_anotherField": 123}
type ExtensionFields map[KindExtension]any

// Set adds or updates an extension field on the entry.
func (e *ExtensionFields) Set(key KindExtension, value any) {
	if *e == nil {
		*e = make(ExtensionFields)
	}
	if len(key) > 0 && key[0] != '_' {
		key = "_" + key
	}
	(*e)[key] = value
}

// Get retrieves a custom extension field by its key.
func (e *ExtensionFields) Get(key KindExtension) (any, bool) {
	if *e == nil {
		return nil, false
	}
	if key[0] != '_' {
		key = "_" + key
	}
	val, ok := (*e)[key]
	return val, ok
}
