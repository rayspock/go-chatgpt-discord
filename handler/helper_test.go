package handler_test

import (
	"github.com/rayspock/go-chatgpt-discord/handler"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSendMessageByChunk(t *testing.T) {
	testCases := map[string]struct {
		message        string
		chunkLength    int
		expectedChunks []string
	}{
		"mandarin": {
			message:        "你好，世界！",
			chunkLength:    2,
			expectedChunks: []string{"你好", "，世", "界！"},
		},
		"english": {
			message:        "Hello World!",
			chunkLength:    2,
			expectedChunks: []string{"He", "ll", "o ", "Wo", "rl", "d!"},
		},
		"japanese": {
			message:        "「こんにちは世界」",
			chunkLength:    2,
			expectedChunks: []string{"「こ", "んに", "ちは", "世界", "」"},
		},
		"spanish": {
			message:        "¡Hola Mundo!",
			chunkLength:    2,
			expectedChunks: []string{"¡H", "ol", "a ", "Mu", "nd", "o!"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			send := make(chan string)
			go handler.SendMessageByChunk(tc.message, tc.chunkLength, send)
			i := 0
			for x := range send {
				assert.Equal(t, tc.expectedChunks[i], x)
				i++
			}
		})
	}
}
