package call_test

// Kiểm thử GET /api/calls/ice-servers: TURN phải được tách thành nhiều
// IceServer (mỗi transport một entry) thay vì gộp thành một URL duy nhất.

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"linkup/config"
	"linkup/controllers"
	"linkup/dto"

	"github.com/gin-gonic/gin"
)

func fetchIceServers(t *testing.T, env config.Env) dto.IceServersResponse {
	t.Helper()
	gin.SetMode(gin.TestMode)

	controller := controllers.NewVoiceCallController(nil, nil, env)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	controller.GetIceServers(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, mong đợi 200", recorder.Code)
	}

	var response dto.IceServersResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("response không phải JSON hợp lệ: %v", err)
	}
	return response
}

func turnEntries(servers []dto.IceServer) []dto.IceServer {
	turns := make([]dto.IceServer, 0, len(servers))
	for _, server := range servers {
		if len(server.URLs) >= 5 && (server.URLs[:5] == "turn:" || server.URLs[:6] == "turns:") {
			turns = append(turns, server)
		}
	}
	return turns
}

func TestGetIceServersWithoutTurnReturnsOnlyStun(t *testing.T) {
	response := fetchIceServers(t, config.Env{
		IceServerUrls: "stun:stun.l.google.com:19302,stun:stun1.l.google.com:19302",
	})

	if len(response.IceServers) != 2 {
		t.Fatalf("mong đợi 2 STUN, nhận %d: %+v", len(response.IceServers), response.IceServers)
	}
	for _, server := range response.IceServers {
		if server.Username != "" || server.Credential != "" {
			t.Fatalf("STUN không được kèm credential: %+v", server)
		}
	}
}

func TestGetIceServersSplitsTurnUrlsPerTransport(t *testing.T) {
	response := fetchIceServers(t, config.Env{
		IceServerUrls: "stun:stun.relay.metered.ca:80",
		TurnServerUrl: "turn:standard.relay.metered.ca:80," +
			"turn:standard.relay.metered.ca:80?transport=tcp," +
			"turn:standard.relay.metered.ca:443," +
			"turns:standard.relay.metered.ca:443?transport=tcp",
		TurnUsername:   "user-1",
		TurnCredential: "pass+with=chars",
	})

	turns := turnEntries(response.IceServers)
	if len(turns) != 4 {
		t.Fatalf("mong đợi 4 entry TURN, nhận %d: %+v", len(turns), turns)
	}
	for _, entry := range turns {
		if entry.Username != "user-1" || entry.Credential != "pass+with=chars" {
			t.Fatalf("credential bị mất hoặc sai: %+v", entry)
		}
	}
	if len(response.IceServers) != 5 {
		t.Fatalf("mong đợi 1 STUN + 4 TURN, nhận %d", len(response.IceServers))
	}
	if turns[3].URLs != "turns:standard.relay.metered.ca:443?transport=tcp" {
		t.Fatalf("URL cuối sai: %q", turns[3].URLs)
	}
}

func TestGetIceServersTrimsAndSkipsEmptySegments(t *testing.T) {
	response := fetchIceServers(t, config.Env{
		TurnServerUrl: " turn:a:3478 ,, turn:b:3478?transport=tcp , ",
		TurnUsername:  "u",
	})

	turns := turnEntries(response.IceServers)
	if len(turns) != 2 {
		t.Fatalf("mong đợi 2 TURN sau khi bỏ đoạn rỗng, nhận %d: %+v", len(turns), turns)
	}
	if turns[0].URLs != "turn:a:3478" {
		t.Fatalf("URL chưa được trim: %q", turns[0].URLs)
	}
	if turns[1].URLs != "turn:b:3478?transport=tcp" {
		t.Fatalf("URL thứ hai sai: %q", turns[1].URLs)
	}
}

func TestGetIceServersEmptyConfigReturnsEmptyList(t *testing.T) {
	response := fetchIceServers(t, config.Env{})

	if len(response.IceServers) != 0 {
		t.Fatalf("mong đợi danh sách rỗng, nhận %+v", response.IceServers)
	}
	if response.IceServers == nil {
		t.Fatal("phải trả mảng rỗng ([]), không phải null")
	}
}

func TestGetIceServersJSONShape(t *testing.T) {
	response := fetchIceServers(t, config.Env{
		TurnServerUrl:  "turn:standard.relay.metered.ca:80",
		TurnUsername:   "user-1",
		TurnCredential: "secret",
	})

	raw, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded map[string][]map[string]string
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	servers, ok := decoded["ice_servers"]
	if !ok || len(servers) != 1 {
		t.Fatalf("thiếu field ice_servers hoặc sai số lượng: %s", raw)
	}
	if servers[0]["urls"] != "turn:standard.relay.metered.ca:80" {
		t.Fatalf("field urls sai: %v", servers[0])
	}
	if servers[0]["username"] != "user-1" || servers[0]["credential"] != "secret" {
		t.Fatalf("credential không được serialize: %v", servers[0])
	}
}
