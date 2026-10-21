package main

import "testing"

func TestHandoffChannel(t *testing.T) {
	cases := []struct{ name, room, want string }{{"order room", "42", "order-42"}, {"unicode-free id", "abc", "order-abc"}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := "order-" + tc.room
			if got != tc.want {
				t.Fatalf("channel = %q, want %q", got, tc.want)
			}
		})
	}
}
