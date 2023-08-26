package handler

import (
	"fmt"
	"unicode/utf8"
)

func SendMessageByChunk(message string, maxChunkLength int, send chan<- string) {
	currentChunkLength := 0
	chunkStartIndex := 0
	prevWordDividerIndex := -1
	for i, w := 0, 0; i < len(message); i += w {
		currentChunkLength++
		char, width := utf8.DecodeRuneInString(message[i:])
		w = width

		// we reach the maximum chunk length
		if currentChunkLength == maxChunkLength {
			chunkToSend, nextChunkStartIndex := getChunkToSend(message, i, chunkStartIndex, prevWordDividerIndex)
			send <- chunkToSend
			chunkStartIndex = nextChunkStartIndex
			currentChunkLength = 0
		} else {
			if IsWordDivider(char) {
				prevWordDividerIndex = i
			}
		}
	}
	if len(message) > 0 && chunkStartIndex < len(message) {
		send <- message[chunkStartIndex:]
	}
	close(send)
}

func getChunkToSend(message string, byteIndex, chunkStartIndex, prevWordDividerIndex int) (string, int) {
	char, runeWidth := utf8.DecodeRuneInString(message[byteIndex:])
	nextChar, _ := utf8.DecodeRuneInString(message[byteIndex+runeWidth:])
	if !IsWordDivider(char) && !IsWordDivider(nextChar) && prevWordDividerIndex != -1 {
		chunkToSend := message[chunkStartIndex : prevWordDividerIndex+1]
		nextChunkStartIndex := prevWordDividerIndex + 1
		return chunkToSend, nextChunkStartIndex
	}
	chunkToSend := message[chunkStartIndex : byteIndex+runeWidth]
	nextChunkStartIndex := byteIndex + runeWidth
	return chunkToSend, nextChunkStartIndex
}

// IsWordDivider returns true if the rune is a word divider.(Currently only space is considered a word divider)
func IsWordDivider(r rune) bool {
	return runeMatch(r, []rune{' '})
}

func getUserMessage(message, author string) string {
	return fmt.Sprintf("> **%s** - <@%s>\n", message, author)
}

func runeMatch(r rune, runes []rune) bool {
	for _, v := range runes {
		if r == v {
			return true
		}
	}
	return false
}
