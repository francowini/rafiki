package priority

import "fmt"

type Priority struct {
	value int
}

func (p Priority) Value() int {
	return p.value
}

func (p Priority) String() string {
	return fmt.Sprintf("%d", p.value)
}

func Parse(value int) (Priority, error) {
	return Priority{value}, nil
}

func MustParse(value int) Priority {
	return Priority{value}
}
