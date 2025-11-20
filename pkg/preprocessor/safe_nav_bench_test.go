package preprocessor

import (
	"testing"
)

// Benchmark data structures for safe navigation tests
type BenchAddress struct {
	Street  string
	City    string
	ZipCode string
}

type BenchUser struct {
	Name    string
	Email   string
	Address *BenchAddress
}

type BenchCompany struct {
	Name string
	CEO  *BenchUser
}

// Setup functions
func getBenchUserWithAddress() *BenchUser {
	return &BenchUser{
		Name:  "John Doe",
		Email: "john@example.com",
		Address: &BenchAddress{
			Street:  "123 Main St",
			City:    "San Francisco",
			ZipCode: "94105",
		},
	}
}

func getBenchUserWithoutAddress() *BenchUser {
	return &BenchUser{
		Name:  "Jane Doe",
		Email: "jane@example.com",
		Address: nil,
	}
}

func getBenchCompanyWithCEO() *BenchCompany {
	return &BenchCompany{
		Name: "Acme Corp",
		CEO:  getBenchUserWithAddress(),
	}
}

// PROPERTY ACCESS BENCHMARKS

// BenchmarkSafeNavPropertyAccess benchmarks user?.address?.city (generated code)
func BenchmarkSafeNavPropertyAccess_WithValue(b *testing.B) {
	user := getBenchUserWithAddress()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Generated code for: user?.address?.city
		result := func() *string {
			if user == nil {
				return nil
			}
			if user.Address == nil {
				return nil
			}
			return &user.Address.City
		}()
		_ = result
	}
}

// BenchmarkHandWrittenNilCheck benchmarks equivalent hand-written code
func BenchmarkHandWrittenNilCheck_WithValue(b *testing.B) {
	user := getBenchUserWithAddress()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var result *string
		if user != nil && user.Address != nil {
			result = &user.Address.City
		}
		_ = result
	}
}

// BenchmarkSafeNavPropertyAccess_NilMiddle benchmarks with nil in chain
func BenchmarkSafeNavPropertyAccess_NilMiddle(b *testing.B) {
	user := getBenchUserWithoutAddress()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Generated code for: user?.address?.city
		result := func() *string {
			if user == nil {
				return nil
			}
			if user.Address == nil {
				return nil
			}
			return &user.Address.City
		}()
		_ = result
	}
}

// BenchmarkHandWrittenNilCheck_NilMiddle benchmarks hand-written with nil
func BenchmarkHandWrittenNilCheck_NilMiddle(b *testing.B) {
	user := getBenchUserWithoutAddress()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var result *string
		if user != nil && user.Address != nil {
			result = &user.Address.City
		}
		_ = result
	}
}

// DEEP CHAIN BENCHMARKS (company?.ceo?.address?.city)

// BenchmarkSafeNavDeepChain_WithValue benchmarks 4-level deep chain
func BenchmarkSafeNavDeepChain_WithValue(b *testing.B) {
	company := getBenchCompanyWithCEO()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Generated code for: company?.ceo?.address?.city
		result := func() *string {
			if company == nil {
				return nil
			}
			if company.CEO == nil {
				return nil
			}
			if company.CEO.Address == nil {
				return nil
			}
			return &company.CEO.Address.City
		}()
		_ = result
	}
}

// BenchmarkHandWrittenDeepChain_WithValue benchmarks hand-written deep chain
func BenchmarkHandWrittenDeepChain_WithValue(b *testing.B) {
	company := getBenchCompanyWithCEO()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var result *string
		if company != nil && company.CEO != nil && company.CEO.Address != nil {
			result = &company.CEO.Address.City
		}
		_ = result
	}
}

// METHOD CALL BENCHMARKS

type BenchUserWithMethods struct {
	Name    string
	Address *BenchAddress
}

func (u *BenchUserWithMethods) GetAddress() *BenchAddress {
	return u.Address
}

func (a *BenchAddress) GetCity() string {
	return a.City
}

func getBenchUserWithMethods() *BenchUserWithMethods {
	return &BenchUserWithMethods{
		Name: "John Doe",
		Address: &BenchAddress{
			City: "San Francisco",
		},
	}
}

// BenchmarkSafeNavMethodChain_WithValue benchmarks user?.getAddress()?.getCity()
func BenchmarkSafeNavMethodChain_WithValue(b *testing.B) {
	user := getBenchUserWithMethods()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Generated code for: user?.getAddress()?.getCity()
		result := func() *string {
			if user == nil {
				return nil
			}
			addr := user.GetAddress()
			if addr == nil {
				return nil
			}
			city := addr.GetCity()
			return &city
		}()
		_ = result
	}
}

// BenchmarkHandWrittenMethodChain_WithValue benchmarks hand-written method chain
func BenchmarkHandWrittenMethodChain_WithValue(b *testing.B) {
	user := getBenchUserWithMethods()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var result *string
		if user != nil {
			addr := user.GetAddress()
			if addr != nil {
				city := addr.GetCity()
				result = &city
			}
		}
		_ = result
	}
}

// SINGLE PROPERTY BENCHMARKS (minimal chain)

// BenchmarkSafeNavSingleProperty_WithValue benchmarks user?.name
func BenchmarkSafeNavSingleProperty_WithValue(b *testing.B) {
	user := getBenchUserWithAddress()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Generated code for: user?.name
		result := func() *string {
			if user == nil {
				return nil
			}
			return &user.Name
		}()
		_ = result
	}
}

// BenchmarkHandWrittenSingleProperty_WithValue benchmarks hand-written user?.name
func BenchmarkHandWrittenSingleProperty_WithValue(b *testing.B) {
	user := getBenchUserWithAddress()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var result *string
		if user != nil {
			result = &user.Name
		}
		_ = result
	}
}

// NIL ROOT BENCHMARKS (test early exit performance)

// BenchmarkSafeNavNilRoot benchmarks with nil root
func BenchmarkSafeNavNilRoot(b *testing.B) {
	var user *BenchUser = nil
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Generated code for: user?.address?.city
		result := func() *string {
			if user == nil {
				return nil
			}
			if user.Address == nil {
				return nil
			}
			return &user.Address.City
		}()
		_ = result
	}
}

// BenchmarkHandWrittenNilRoot benchmarks hand-written with nil root
func BenchmarkHandWrittenNilRoot(b *testing.B) {
	var user *BenchUser = nil
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var result *string
		if user != nil && user.Address != nil {
			result = &user.Address.City
		}
		_ = result
	}
}

// IIFE OPTIMIZATION VALIDATION
// These benchmarks verify that Go compiler inlines IIFEs effectively

// BenchmarkIIFEOverhead_Minimal benchmarks minimal IIFE
func BenchmarkIIFEOverhead_Minimal(b *testing.B) {
	value := 42
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result := func() int {
			return value
		}()
		_ = result
	}
}

// BenchmarkDirectAccess_Minimal benchmarks direct access (no IIFE)
func BenchmarkDirectAccess_Minimal(b *testing.B) {
	value := 42
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result := value
		_ = result
	}
}

// BenchmarkIIFEOverhead_WithCondition benchmarks IIFE with conditional
func BenchmarkIIFEOverhead_WithCondition(b *testing.B) {
	value := 42
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result := func() int {
			if value > 0 {
				return value
			}
			return 0
		}()
		_ = result
	}
}

// BenchmarkDirectCondition benchmarks direct conditional
func BenchmarkDirectCondition(b *testing.B) {
	value := 42
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var result int
		if value > 0 {
			result = value
		} else {
			result = 0
		}
		_ = result
	}
}
