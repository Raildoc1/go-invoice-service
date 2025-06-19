package flagtypes

import (
	"flag"
	"strconv"
)

const nilString = "<nil>"

// String.

var _ flag.Value = (*String)(nil)

type String struct {
	Argument[string]
}

func NewString() *String {
	return &String{
		*newArgument[string](),
	}
}

func (s *String) String() string {
	if s.val == nil {
		return nilString
	}
	return *s.val
}

func (s *String) Set(input string) error {
	s.val = &input
	return nil
}

// Int.

var _ flag.Value = (*Int)(nil)

type Int struct {
	Argument[int]
}

func NewInt() *Int {
	return &Int{
		*newArgument[int](),
	}
}

func (i *Int) String() string {
	if i.val == nil {
		return nilString
	}
	return strconv.Itoa(*i.val)
}

func (i *Int) Set(input string) error {
	val, err := strconv.Atoi(input)
	if err != nil {
		return err
	}
	i.val = &val
	return nil
}

// Int.

var _ flag.Value = (*Bool)(nil)

type Bool struct {
	Argument[bool]
}

func NewBool() *Bool {
	return &Bool{
		*newArgument[bool](),
	}
}

func (b *Bool) String() string {
	if b.val == nil {
		return nilString
	}
	return strconv.FormatBool(*b.val)
}

func (b *Bool) Set(input string) error {
	val, err := strconv.ParseBool(input)
	if err != nil {
		return err
	}
	b.val = &val
	return nil
}
