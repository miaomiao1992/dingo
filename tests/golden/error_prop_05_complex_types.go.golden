package main

type User struct {
	ID   int
	Name string
}

func fetchUser(id int) (*User, error) {
	__tmp0, __err0 := ReadFile("user.json")
	// dingo:s:1
	if __err0 != nil {
		return nil, __err0
	}
	// dingo:e:1
	var data = __tmp0
	return &User{ID: id, Name: string(data)}, nil
}
func getNames() ([]string, error) {
	__tmp0, __err0 := ReadFile("names.txt")
	// dingo:s:1
	if __err0 != nil {
		return nil, __err0
	}
	// dingo:e:1
	var data = __tmp0
	return []string{string(data)}, nil
}
