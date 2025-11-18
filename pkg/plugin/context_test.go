package plugin

import (
	"go/token"
	"testing"
)

func TestContext_ReportError(t *testing.T) {
	ctx := &Context{}

	// Initially, no errors
	if ctx.HasErrors() {
		t.Error("HasErrors() should be false initially")
	}

	// Report an error
	ctx.ReportError("test error 1", token.Pos(10))

	if !ctx.HasErrors() {
		t.Error("HasErrors() should be true after reporting error")
	}

	errors := ctx.GetErrors()
	if len(errors) != 1 {
		t.Errorf("GetErrors() returned %d errors, want 1", len(errors))
	}

	// Report another error
	ctx.ReportError("test error 2", token.Pos(20))

	errors = ctx.GetErrors()
	if len(errors) != 2 {
		t.Errorf("GetErrors() returned %d errors, want 2", len(errors))
	}
}

func TestContext_GetErrors_Empty(t *testing.T) {
	ctx := &Context{}

	errors := ctx.GetErrors()
	if errors == nil {
		t.Error("GetErrors() should return empty slice, not nil")
	}
	if len(errors) != 0 {
		t.Errorf("GetErrors() returned %d errors, want 0", len(errors))
	}
}

func TestContext_ClearErrors(t *testing.T) {
	ctx := &Context{}

	// Add some errors
	ctx.ReportError("error 1", token.Pos(10))
	ctx.ReportError("error 2", token.Pos(20))

	if !ctx.HasErrors() {
		t.Error("HasErrors() should be true")
	}

	// Clear errors
	ctx.ClearErrors()

	if ctx.HasErrors() {
		t.Error("HasErrors() should be false after ClearErrors()")
	}

	errors := ctx.GetErrors()
	if len(errors) != 0 {
		t.Errorf("GetErrors() returned %d errors after clear, want 0", len(errors))
	}
}

func TestContext_NextTempVar(t *testing.T) {
	ctx := &Context{}

	// First call should return __tmp0
	name1 := ctx.NextTempVar()
	if name1 != "__tmp0" {
		t.Errorf("NextTempVar() = %q, want %q", name1, "__tmp0")
	}

	// Counter should increment
	if ctx.TempVarCounter != 1 {
		t.Errorf("TempVarCounter = %d, want 1", ctx.TempVarCounter)
	}

	// Second call should return __tmp1
	name2 := ctx.NextTempVar()
	if name2 != "__tmp1" {
		t.Errorf("NextTempVar() = %q, want %q", name2, "__tmp1")
	}

	// Third call should return __tmp2
	name3 := ctx.NextTempVar()
	if name3 != "__tmp2" {
		t.Errorf("NextTempVar() = %q, want %q", name3, "__tmp2")
	}

	// Counter should be 3
	if ctx.TempVarCounter != 3 {
		t.Errorf("TempVarCounter = %d, want 3", ctx.TempVarCounter)
	}
}

func TestContext_NextTempVar_UniqueNames(t *testing.T) {
	ctx := &Context{}

	// Generate 10 temp var names
	names := make(map[string]bool)
	for i := 0; i < 10; i++ {
		name := ctx.NextTempVar()
		if names[name] {
			t.Errorf("Duplicate temp var name: %s", name)
		}
		names[name] = true
	}

	// Should have 10 unique names
	if len(names) != 10 {
		t.Errorf("Generated %d unique names, want 10", len(names))
	}
}

func TestContext_ErrorsWithLocation(t *testing.T) {
	ctx := &Context{}

	pos1 := token.Pos(100)
	pos2 := token.Pos(200)

	ctx.ReportError("error at pos 100", pos1)
	ctx.ReportError("error at pos 200", pos2)

	errors := ctx.GetErrors()
	if len(errors) != 2 {
		t.Fatalf("Expected 2 errors, got %d", len(errors))
	}

	// Check that error messages contain position info
	err1Str := errors[0].Error()
	if !contains(err1Str, "100") {
		t.Errorf("Error message missing position 100: %s", err1Str)
	}

	err2Str := errors[1].Error()
	if !contains(err2Str, "200") {
		t.Errorf("Error message missing position 200: %s", err2Str)
	}
}

// Helper function to check substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
