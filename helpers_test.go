package singlepage

import "testing"

func TestIsRemote(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"./asd", false},
		{"http://./asd", true},
		{"//./asd", true},
		{"://./asd", false},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			out := isRemote(tt.in)
			if out != tt.want {
				t.Errorf("\nout:  %#v\nwant: %#v\n", out, tt.want)
			}
		})
	}
}
