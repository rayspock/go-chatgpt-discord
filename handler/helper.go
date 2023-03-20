package handler

import "fmt"

func getUserMessage(message, author string) string {
	return fmt.Sprintf("> **%s** - <@%s>", message, author)
}
