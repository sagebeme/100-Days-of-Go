package discordbot

import (
	"crypto/ed25519"
	"net/http"
)

// The headers Discord signs every interaction request with.
const (
	SignatureHeader = "X-Signature-Ed25519"
	TimestampHeader = "X-Signature-Timestamp"
)

// Interaction and response types from Discord's API.
const (
	InteractionPing               = 1
	InteractionApplicationCommand = 2

	ResponsePong                     = 1
	ResponseChannelMessageWithSource = 4

	FlagEphemeral = 64
)

// MaxBodyBytes caps an interaction request.
const MaxBodyBytes = 64 << 10

// VerifyRequest checks Discord's signature: an Ed25519 signature, hex-encoded in
// sigHex, over the timestamp followed by the body. It returns false for anything
// malformed, and never panics.
func VerifyRequest(publicKey ed25519.PublicKey, sigHex, timestamp string, body []byte) bool {
	// Your code here
	return false
}

// ParseCommand turns an APPLICATION_COMMAND interaction (type 2) body into a Command.
// The name is data.name; each of data.options becomes Options[name] = value as text
// (the number 93 becomes "93"); the user ID comes from member.user.id in a server or
// user.id in a DM. Options is never nil. Any other type is an error.
func ParseCommand(body []byte) (Command, error) {
	// Your code here
	return Command{}, nil
}

// NewInteractionsHandler returns the handler for the "Interactions Endpoint URL" in
// Discord's developer portal:
//
//	405  not a POST
//	401  bad or missing signature (Discord tests this before it saves your URL)
//	400  a body that isn't JSON, or an interaction that isn't a PING or a command
//	200  {"type":1} for a PING
//	200  {"type":4,"data":{"content":"...","flags":64}} for a command; flags only when ephemeral
func NewInteractionsHandler(publicKey ed25519.PublicKey, h Handler) http.Handler {
	// Your code here
	return http.NotFoundHandler()
}
