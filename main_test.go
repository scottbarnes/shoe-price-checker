package main

import (
	"reflect"
	"testing"
)

func TestGetShoesAtOrBelowThreshold(t *testing.T) {
	type test struct {
		name      string
		threshold float64
		items     []ShoeItem
		want      []ShoeItem
		err       error
	}

	tests := []test{
		{
			name:      "Threshold without decimals includes items that have decimals, when values are the same",
			threshold: 50,
			items: []ShoeItem{
				{
					"Merrel Trail Glove",
					"",
					89.99,
					79.99,
				}, {
					"Hoka Challenger",
					"Blah",
					160.99,
					50,
				}, {
					"Altra Lone Peak",
					"",
					120.99,
					45.99,
				},
			},
			want: []ShoeItem{
				{
					"Altra Lone Peak",
					"",
					120.99,
					45.99,
				}, {
					"Hoka Challenger",
					"Blah",
					160.99,
					50.00,
				},
			},
			err: nil,
		}, {
			name:      "Threshold with decimals includes items lack decimals, when values are the same",
			threshold: 50.00,
			items: []ShoeItem{
				{
					"Merrel Trail Glove",
					"",
					89.99,
					79.99,
				}, {
					"Hoka Challenger",
					"Blah",
					160.99,
					50,
				}, {
					"Altra Lone Peak",
					"",
					120.99,
					45.99,
				},
			},
			want: []ShoeItem{
				{
					"Altra Lone Peak",
					"",
					120.99,
					45.99,
				}, {
					"Hoka Challenger",
					"Blah",
					160.99,
					50,
				},
			},
			err: nil,
		}, {
			name:      "Handles empty []ShoeItem{}",
			threshold: 55.43,
			items:     []ShoeItem{},
			want:      []ShoeItem{},
			err:       nil,
		},
	}

	for _, tc := range tests {
		got := getShoesAtOrBelowThreshold(tc.items, tc.threshold)
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("expected: %v, got %v", tc.want, got)
		}
	}
}
