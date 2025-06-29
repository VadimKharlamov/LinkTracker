package utils_test

import (
	"bot/internal/model/bot"
	"bot/utils"

	"reflect"
	"testing"
)

func TestJSONCodec_Marshal(t *testing.T) {
	codec := utils.JSONCodec{}

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name: "valid struct",
			input: bot.LinkUpdate{
				ID:          123,
				URL:         "https://example.com",
				Description: "desc",
				TgChatIDs:   []int64{1, 2, 3},
			},
			wantErr: false,
		},
		{
			name:    "unsupported type",
			input:   make(chan int),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := codec.Marshal(tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("got error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestJSONCodec_Unmarshal(t *testing.T) {
	codec := utils.JSONCodec{}

	tests := []struct {
		name    string
		data    []byte
		want    *bot.LinkUpdate
		wantErr bool
	}{
		{
			name:    "valid JSON",
			data:    []byte(`{"id":123,"url":"https://example.com","description":"desc","tgChatIds":[1,2,3]}`),
			want:    &bot.LinkUpdate{ID: 123, URL: "https://example.com", Description: "desc", TgChatIDs: []int64{1, 2, 3}},
			wantErr: false,
		},
		{
			name:    "missing required field",
			data:    []byte(`{"url":"no id","description":"desc","tgChatIds":[1]}`),
			want:    &bot.LinkUpdate{URL: "no id", Description: "desc", TgChatIDs: []int64{1}},
			wantErr: false,
		},
		{
			name:    "invalid JSON syntax",
			data:    []byte(`{"id":not-an-int}`),
			want:    nil,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var got *bot.LinkUpdate

			err := codec.Unmarshal(tc.data, &got)

			if (err != nil) != tc.wantErr {
				t.Errorf("got error = %v, wantErr %v", err, tc.wantErr)
			}

			if err == nil && !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got error = %+v, want %+v", got, *tc.want)
			}
		})
	}
}
