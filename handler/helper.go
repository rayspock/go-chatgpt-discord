package handler

import (
	"fmt"
	"unicode/utf8"
)

func getUserMessage(message, author string) string {
	return fmt.Sprintf("> **%s** - <@%s>\n", message, author)
}

func SendMessageByChunk(message string, chunkLength int, send chan<- string) {
	counter := 0
	lastIndex := 0
	for i, w := 0, 0; i < len(message); i += w {
		counter++
		_, width := utf8.DecodeRuneInString(message[i:])
		w = width

		// we reach the maximum chunk length
		if counter == chunkLength {
			chunk := message[lastIndex : i+w]
			send <- chunk
			lastIndex = i + w
			counter = 0
		}
	}
	if len(message) > 0 && lastIndex < len(message) {
		send <- message[lastIndex:]
	}
	close(send)
}
