package mock

// MockDBHandler is a mock implementation of DBHandler for testing
type MockDBHandler struct {
	CreateFunc func(db string, data any) (string, error)
	FindFunc   func(db string, filter map[string]any) ([]any, error)
	UpdateFunc func(db string, filter map[string]any, data any) error
	DeleteFunc func(db string, filter map[string]any) error
}

func (m *MockDBHandler) Create(db string, data any) (string, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(db, data)
	}
	return "mock-id", nil
}

func (m *MockDBHandler) Find(db string, filter map[string]any) ([]any, error) {
	if m.FindFunc != nil {
		return m.FindFunc(db, filter)
	}
	return []any{}, nil
}

func (m *MockDBHandler) Update(db string, filter map[string]any, data any) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(db, filter, data)
	}
	return nil
}

func (m *MockDBHandler) Delete(db string, filter map[string]any) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(db, filter)
	}
	return nil
}
