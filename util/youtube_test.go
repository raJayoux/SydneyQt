package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestYoutube(t *testing.T) {
	for _, link := range []string{"https://www.youtube.com/watch?v=LUch7N9Gw28",
		"https://www.youtube.com/watch?v=yqHLW60BSlA"} {
		yt, _ := NewYoutube(link, "")
		vd, err := yt.GetVideoDetails()
		assert.Nil(t, err)
		assert.NotEmpty(t, vd.Title)
		c, err := yt.GetCaptions()
		assert.Nil(t, err)
		assert.NotEmpty(t, c)
		tx, err := c[0].GetTranscript("")
		assert.Nil(t, err)
		assert.NotEmpty(t, tx)
	}
}
