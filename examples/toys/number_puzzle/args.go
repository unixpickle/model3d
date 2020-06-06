package main

import (
	"flag"
)

type Args struct {
	SegmentThickness   float64
	SegmentDepth       float64
	SegmentTipInset    float64
	SegmentJointOutset float64

	BoardThickness float64
	BoardBorder    float64
}

func (a *Args) Add() {
	flag.Float64Var(&a.SegmentThickness, "segment-thickness", 0.15,
		"horizontal thickness of digits")
	flag.Float64Var(&a.SegmentDepth, "segment-depth", 0.1, "depthwise thickness of digits")
	flag.Float64Var(&a.SegmentTipInset, "segment-tip-inset", 0.03, "slack at unconnected tips")
	flag.Float64Var(&a.SegmentJointOutset, "segment-join-outset", 0.015,
		"extra length to connect segments")
	flag.Float64Var(&a.BoardThickness, "board-thickness", 0.3, "thickness of board base")
	flag.Float64Var(&a.BoardBorder, "board-border", 0.2, "border around the board")
}
