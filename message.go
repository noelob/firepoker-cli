package main

import (
	"encoding/json"
	"iter"
	"log/slog"
	"maps"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	// ErrUnknownMessage is returned if the message is unknown
	ErrUnknownMessage = Error("unable to handle message")
)

var (
	participantPattern = regexp.MustCompile(`^games/[a-f0-9\-]+/participants/[a-f0-9\-]+$`)
	onlinePattern      = regexp.MustCompile(`^games/[a-f0-9\-]+/participants/[a-f0-9\-]+/online$`)
	gamePattern        = regexp.MustCompile(`^games/[a-f0-9\-]+$`)
)

type User struct {
	Id       string `json:"id"`
	FullName string `json:"fullname"`
	HasVoted bool   `json:"hasVoted"`
	//Online   time.Time `json:"online"`
}

type Presence struct {
	Id       string
	Online   bool
	LastSeen time.Time
}

type Result struct {
	Points uint16 `json:"points"`
	User   User   `json:"user"`
}

type Estimate struct {
	Id      uint16            `json:"id"`
	Title   string            `json:"title"`
	Status  string            `json:"status"`
	Results map[string]Result `json:"results"`
}

type Story struct {
	Id     uint16 `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

type GameState struct {
	//Created      time.Time        `json:"created"`
	Deck         uint             `json:"deck"`
	Description  string           `json:"description"`
	Estimate     Estimate         `json:"estimate"`
	Name         string           `json:"name"`
	Owner        User             `json:"owner"`
	Participants map[string]User  `json:"participants"`
	Status       string           `json:"status"`
	Stories      map[string]Story `json:"stories"`
}

type Body struct {
	P string          `json:"p"`
	D json.RawMessage `json:"d"`
	S string          `json:"s"`
	//D BodyData `json:"d"`
}

type Acknowledgement struct {
	Ref uint16
	S   string `json:"s"`
}

type DataFrame struct {
	Type string          `json:"t"`
	Ref  uint16          `json:"r"`
	Data json.RawMessage `json:"d"`
	Body json.RawMessage `json:"b"`
	//Data InnerData `json:"d"`
	//Body Body      `json:"b"`
}

type Handshake struct {
	Timestamp int64  `json:"ts"`
	Version   string `json:"v"`
	Host      string `json:"h"`
	SessionID string `json:"s"`
}

type ControlFrame struct {
	Type string          `json:"t"`
	Data json.RawMessage `json:"d"`
}

type Frame struct {
	Type string          `json:"t"`
	Data json.RawMessage `json:"d"`
}

// ParseMessage parses arbitrary websocket messages and returns the correct message type,
func ParseMessage(data []byte) (any, error) {
	// 1. Control or Data frame
	var frame Frame
	err := json.Unmarshal(data, &frame)
	if err != nil {
		slog.Warn("<<< Unable to parse frame: %v", err)
	} else {
		switch frame.Type {
		case "c": // control
			return parseControlFrame(frame.Data)
		case "d": //data
			return parseDataFrame(frame.Data)
		}
	}

	return nil, ErrUnknownMessage

}

func parseControlFrame(data []byte) (any, error) {
	slog.Debug("<<< this is a control frame\n")
	var ctlFrame ControlFrame
	//fmt.Printf("<<< Raw JSON: '%s'\n", string(frame.Data))
	err := json.Unmarshal(data, &ctlFrame)
	if err != nil {
		slog.Error("<<< Unable to parse control frane: %v", err)
		return nil, err
	}

	switch ctlFrame.Type {
	case "h": //handshake
		slog.Debug("<<<< this is a handshake")
		var handshake Handshake
		err = json.Unmarshal(ctlFrame.Data, &handshake)
		if err != nil {
			slog.Warn("<<<< Unable to parse handshake: %v", err)
			return nil, err
		}

		return handshake, nil
	}

	return nil, ErrUnknownMessage
}

func parseDataFrame(data []byte) (any, error) {
	slog.Debug("<<< this is a data frame")
	var dataFrame DataFrame
	err := json.Unmarshal(data, &dataFrame)
	if err != nil {
		slog.Error("<<< Unable to parse data frame: %v", err)
		return nil, err
	}

	// 1. Acknowledgement message
	if dataFrame.Ref > 0 {
		ack := Acknowledgement{Ref: dataFrame.Ref}
		err := json.Unmarshal(dataFrame.Body, &ack)
		if err != nil {
			slog.Error("<<< Unable to parse ack frame: %v", err)
			return nil, err
		} else {
			return ack, nil
		}
	}

	// Need to check if there's a body?
	if dataFrame.Data != nil {
		slog.Info("THERE IS INNER FRAME DATA")
	}

	if dataFrame.Body != nil {
		var body Body
		err := json.Unmarshal(dataFrame.Body, &body)
		if err != nil {
			slog.Error("<<< Unable to parse body: %v", err)
			return nil, err
		}

		switch prop := body.P; {
		// 2. Participant
		// "games/d2538816-2f8e-a8b0-6534-30857b5e932d/participants/31d8c788-0105-7854-e577-f08aa28a9024"
		case participantPattern.MatchString(prop):
			return parseParticipant(body.D)

		// 3. Participant Online
		// "games/d2538816-2f8e-a8b0-6534-30857b5e932d/participants/31d8c788-0105-7854-e577-f08aa28a9024/online"
		case onlinePattern.MatchString(prop):
			return parsePresence(body.P, body.D)

		// 4. Game message
		// "games/d2538816-2f8e-a8b0-6534-30857b5e932d"
		case gamePattern.MatchString(prop):
			return parseGame(body.D)
		}
	}

	return nil, ErrUnknownMessage
}

func parseParticipant(data []byte) (User, error) {
	var user User
	err := json.Unmarshal(data, &user)
	if err != nil {
		slog.Error("<<< Unable to parse user: %v", err)
		return User{}, err
	}
	return user, nil
}

func parsePresence(path string, data []byte) (Presence, error) {
	parts := strings.Split(path, "/")

	presence := Presence{
		Id: parts[3],
	}

	// The data can be either a boolean (true) or an epoch timestamp indicating the last seen time :/
	a := string(data)

	if a == "true" {
		presence.Online = true
	} else {
		ts, err := strconv.Atoi(a)
		if err != nil {
			slog.Error("<<< Unable to parse presence timestamp: %v", err)
		} else {
			presence.LastSeen = time.Unix(int64(ts), 0)
		}
	}

	return presence, nil
}

func parseGame(data []byte) (any, error) {
	// 1. Game state
	var gameState GameState
	err := json.Unmarshal(data, &gameState)
	if err != nil {
		slog.Error("<<< Unable to parse gameState struct: %v", err)
	}

	if !reflect.DeepEqual(gameState, GameState{}) {
		return gameState, nil
	}

	// 2. participant offline
	// The JSON key is a RESTful path :/
	var props map[string]int64

	// Unmarshal the JSON string (converted to byte slice) into the map
	err = json.Unmarshal(data, &props)
	if err != nil {
		slog.Error("<<< Unable to parse gameState map", err)
		return GameState{}, err
	}

	if len(props) == 1 {
		// {"participants/ff7f32af-455c-45e1-601d-79e6488d6887/online":1755365240769}}
		pairs := maps.All(props)
		next, stop := iter.Pull2(pairs)
		defer stop()
		k, v, _ := next()
		parts := strings.Split(k, "/")
		return Presence{
			Id:       parts[1],
			LastSeen: time.Unix(v, 0),
		}, nil
	}

	return nil, ErrUnknownMessage
}
