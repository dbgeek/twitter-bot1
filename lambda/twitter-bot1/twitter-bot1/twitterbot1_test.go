package main

import (
	"fmt"
	"testing"
)

func TestTwitterCrcCheck(t *testing.T) {
	tt := []struct {
		name            string
		twitterCrcToken string
		consumerSecret  string
		out             string
	}{
		{
			name:            "simpleUnitTestforTwitterCrcToken",
			twitterCrcToken: "helloWorld",
			consumerSecret:  "ss",
			out:             "sha256=YFpPr1o5UmzuIUDQn+BqYQ14kFOjWiWYc7oNiVymMgg=",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			consumerSecret = tc.consumerSecret
			result := newCrsToken(tc.twitterCrcToken)
			if tc.out != fmt.Sprintf("%v", result) {
				t.Fatalf("got: %v, wanted: %v", result, tc.out)
			}
		})
	}
}
