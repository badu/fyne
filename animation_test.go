package fyne

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnimationLinear(t *testing.T) {
	assert.Equal(t, float32(0.25), AnimationLinear(0.25))
}

func TestAnimationEaseInOut(t *testing.T) {
	assert.Equal(t, float32(0.125), AnimationEaseInOut(0.25))
	assert.Equal(t, float32(0.5), AnimationEaseInOut(0.5))
	assert.Equal(t, float32(0.875), AnimationEaseInOut(0.75))
}

func TestAnimationEaseIn(t *testing.T) {
	assert.Equal(t, float32(0.0625), AnimationEaseIn(0.25))
	assert.Equal(t, float32(0.25), AnimationEaseIn(0.5))
	assert.Equal(t, float32(0.5625), AnimationEaseIn(0.75))
}

func TestAnimationEaseOut(t *testing.T) {
	assert.Equal(t, float32(0.4375), AnimationEaseOut(0.25))
	assert.Equal(t, float32(0.75), AnimationEaseOut(0.5))
	assert.Equal(t, float32(0.9375), AnimationEaseOut(0.75))
}
