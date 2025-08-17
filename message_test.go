package main

import (
	"io"
	"os"
	"reflect"
	"testing"
	"time"
)

type testCase struct {
	messageFile string
	want        any
	wantErr     bool
}

func TestParseMessage_Success(t *testing.T) {
	tests := map[string]testCase{
		"handshake": {
			messageFile: "testdata/responses/00-handshake.json",
			want: Handshake{
				Timestamp: 1754538609500,
				Version:   "5",
				Host:      "s-usc1b-nss-2107.firebaseio.com",
				SessionID: "8BOJ3puMQW7hbsaOWidnRG6mmc4BOLdl",
			},
			wantErr: false,
		},
		"acknowledgement": {
			messageFile: "testdata/responses/01-join-01-status.json",
			want: Acknowledgement{
				Ref: 1,
				S:   "ok",
			},
			wantErr: false,
		},
		"participant joins": {
			messageFile: "testdata/responses/20-new-participant-01-participant-details.json",
			want: User{
				Id:       "ff7f32af-455c-45e1-601d-79e6488d6887",
				FullName: "John Doe",
				HasVoted: false,
			},
			wantErr: false,
		},
		"participant online": {
			messageFile: "testdata/responses/20-new-participant-02-participant-online.json",
			want: Presence{
				Id:       "31d8c788-0105-7854-e577-f08aa28a9024",
				Online:   true,
				LastSeen: time.Time{},
			},
			wantErr: false,
		},
		"participant offline": {
			messageFile: "testdata/responses/10-refresh-01-participant-offline.json",
			want: Presence{
				Id:       "31d8c788-0105-7854-e577-f08aa28a9024",
				Online:   false,
				LastSeen: time.Unix(int64(1755217636141), 0),
			},
			wantErr: false,
		},
		"game state": {
			messageFile: "testdata/responses/01-join-01-game-state.json",
			want: GameState{
				Deck:        0,
				Description: "The Game Description",
				Estimate: Estimate{
					Id:     1,
					Title:  "Story 2",
					Status: "active",
					Results: map[string]Result{
						"0": {
							Points: 8,
							User: User{
								Id:       "b01f4c74-9eef-86aa-ac09-02a727ca1d31",
								FullName: "John",
							},
						},
					},
				},
				Name: "The Game Name",
				Owner: User{
					Id:       "31d8c788-0105-7854-e577-f08aa28a9024",
					FullName: "Admin",
				},
				Participants: map[string]User{
					"31d8c788-0105-7854-e577-f08aa28a9024": {
						Id:       "31d8c788-0105-7854-e577-f08aa28a9024",
						FullName: "Admin",
					},
					"b01f4c74-9eef-86aa-ac09-02a727ca1d31": {
						Id:       "b01f4c74-9eef-86aa-ac09-02a727ca1d31",
						FullName: "John",
						HasVoted: true,
					},
					"fe24478e-0161-0c97-18ef-ab569207ac44": {
						Id:       "fe24478e-0161-0c97-18ef-ab569207ac44",
						FullName: "go-cli",
					},
				},
				Status: "active",
				Stories: map[string]Story{
					"0": {
						Id:     0,
						Title:  "Story 1",
						Status: "queue",
					},
					"1": {
						Id:     1,
						Title:  "Story 2",
						Status: "active",
					},
					"2": {
						Id:     0,
						Title:  "Story 2",
						Status: "queue",
					},
				},
			},
			wantErr: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			message, err := loadMessage(test.messageFile)
			if err != nil {
				t.Fatalf("unable to read file '%s': %s", test.messageFile, err.Error())
			}

			got, err := ParseMessage(message)
			if err == nil && test.wantErr {
				t.Fatalf("expected an error %s, got none", err.Error())
			}

			if err != nil && !test.wantErr {
				t.Fatalf("expected no error, got one %s", err.Error())
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("differnet result: got %v, wanted %v", got, test.want)
			}
		})
	}
}

func loadMessage(filePath string) ([]byte, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return io.ReadAll(f)

}
