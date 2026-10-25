package discordbot

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
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
// malformed.
func VerifyRequest(publicKey ed25519.PublicKey, sigHex, timestamp string, body []byte) bool {
	if len(publicKey) != ed25519.PublicKeySize || timestamp == "" {
		return false
	}
	sig, err := hex.DecodeString(sigHex)
	if err != nil || len(sig) != ed25519.SignatureSize {
		return false
	}
	msg := make([]byte, 0, len(timestamp)+len(body))
	msg = append(msg, timestamp...)
	msg = append(msg, body...)
	return ed25519.Verify(publicKey, msg, sig)
}

// interaction is the part of Discord's interaction object the bot reads.
type interaction struct {
	Type int `json:"type"`
	Data struct {
		Name    string `json:"name"`
		Options []struct {
			Name  string `json:"name"`
			Value any    `json:"value"` // a string, number or bool
		} `json:"options"`
	} `json:"data"`
	Member *struct {
		User struct {
			ID string `json:"id"`
		} `json:"user"`
	} `json:"member"` // set in a server
	User *struct {
		ID string `json:"id"`
	} `json:"user"` // set in a DM
}

type interactionResponse struct {
	Type int          `json:"type"`
	Data *messageData `json:"data,omitempty"`
}

type messageData struct {
	Content string `json:"content"`
	Flags   int    `json:"flags,omitempty"`
}

// ParseCommand turns an APPLICATION_COMMAND interaction body into a Command. The user
// ID comes from member.user (in a server) or user (in a DM).
func ParseCommand(body []byte) (Command, error) {
	var in interaction
	if err := json.Unmarshal(body, &in); err != nil {
		return Command{}, err
	}
	if in.Type != InteractionApplicationCommand {
		return Command{}, fmt.Errorf("interaction type %d is not a command", in.Type)
	}
	c := Command{Name: in.Data.Name, Options: map[string]string{}}
	for _, o := range in.Data.Options {
		c.Options[o.Name] = optionText(o.Value)
	}
	switch {
	case in.Member != nil:
		c.UserID = in.Member.User.ID
	case in.User != nil:
		c.UserID = in.User.ID
	}
	return c, nil
}

// optionText writes an option value as text. JSON numbers arrive as float64, so 93
// must become "93", not "93.000000".
func optionText(v any) string {
	switch v := v.(type) {
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	}
	return ""
}

// NewInteractionsHandler returns the handler for the "Interactions Endpoint URL" in
// Discord's developer portal. It answers 405 for anything but POST, 401 for a bad
// signature, PONG to a PING, and runs every command through h.
func NewInteractionsHandler(publicKey ed25519.PublicKey, h Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "use POST", http.StatusMethodNotAllowed)
			return
		}
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, MaxBodyBytes))
		if err != nil {
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}
		if !VerifyRequest(publicKey, r.Header.Get(SignatureHeader), r.Header.Get(TimestampHeader), body) {
			http.Error(w, "invalid request signature", http.StatusUnauthorized)
			return
		}
		var probe struct {
			Type int `json:"type"`
		}
		if err := json.Unmarshal(body, &probe); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if probe.Type == InteractionPing {
			writeJSON(w, interactionResponse{Type: ResponsePong})
			return
		}
		c, err := ParseCommand(body)
		if err != nil {
			http.Error(w, "unsupported interaction", http.StatusBadRequest)
			return
		}
		resp := h.Handle(r.Context(), c)
		data := &messageData{Content: resp.Content}
		if resp.Ephemeral {
			data.Flags = FlagEphemeral
		}
		writeJSON(w, interactionResponse{Type: ResponseChannelMessageWithSource, Data: data})
	})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
