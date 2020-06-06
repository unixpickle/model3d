package main

import "sort"

// FixedTemplateNames returns the keys from
// FixedTemplates(), sorted by name.
func FixedTemplateNames() []string {
	var res []string
	for k := range FixedTemplates() {
		res = append(res, k)
	}
	sort.Strings(res)
	return res
}

// FixedTemplates gets a collection of fixed segment
// patterns for constraining puzzles.
func FixedTemplates() map[string][]Digit {
	return map[string][]Digit{
		"none": []Digit{},
		"middle-square": []Digit{
			NewDigitContinuous([]Location{
				{2, 2},
				{2, 3},
				{3, 3},
				{3, 2},
				{2, 2},
			}),
		},
		"three-squares": []Digit{
			NewDigitContinuous([]Location{
				{2, 2},
				{2, 3},
				{3, 3},
				{3, 2},
				{2, 2},
			}),
			NewDigitContinuous([]Location{
				{0, 0},
				{0, 1},
				{1, 1},
				{1, 0},
				{0, 0},
			}),
			NewDigitContinuous([]Location{
				{4, 4},
				{4, 5},
				{5, 5},
				{5, 4},
				{4, 4},
			}),
		},
		"diagonal": []Digit{
			NewDigitContinuous([]Location{
				{0, 0},
				{0, 1},
				{1, 1},
				{1, 2},
				{2, 2},
				{2, 3},
				{3, 3},
				{3, 4},
				{4, 4},
				{4, 5},
				{5, 5},
			}),
		},
	}
}
