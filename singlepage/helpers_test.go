package singlepage

import "testing"

func TestIsRemote(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"./asd", false},
		{"http://./asd", true},
		{"//./asd", true},
		{"://./asd", false},
	}

	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			out := isRemote(tc.in)
			if out != tc.want {
				t.Errorf("\nout:  %#v\nwant: %#v\n", out, tc.want)
			}
		})
	}
}
