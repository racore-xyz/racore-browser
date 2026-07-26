package localai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestChatParsesLocalBrowserPlan(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/health":
			w.WriteHeader(http.StatusOK)
		case "/completion":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"content":"[{\"name\":\"search_web\",\"arguments\":{\"query\":\"Racore browser\"}}]"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	manager := NewClient(server.URL, server.Client())
	response, err := manager.Chat(
		context.Background(),
		[]map[string]string{{"role": "user", "content": "search for Racore browser"}},
		"",
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Actions) != 1 {
		t.Fatalf("expected one action, got %#v", response.Actions)
	}
	if response.Actions[0].Type != "search_web" || response.Actions[0].Value != "Racore browser" {
		t.Fatalf("unexpected action: %#v", response.Actions[0])
	}
	if response.Engine != Engine || response.Model != ModelID {
		t.Fatalf("unexpected local model identity: %#v", response)
	}
}

func TestPromptContainsOnlyLocalBrowserTools(t *testing.T) {
	prompt := buildPrompt(
		[]map[string]string{{"role": "user", "content": "open example.com"}},
		"Ask before side effects.",
	)
	for _, tool := range []string{
		"answer_user",
		"ask_user",
		"open_url",
		"search_web",
		"resolve_racore_domain",
	} {
		if !strings.Contains(prompt, tool) {
			t.Fatalf("prompt does not contain %s", tool)
		}
	}
	if strings.Contains(strings.ToLower(prompt), "openai") {
		t.Fatal("prompt must not reference a third-party provider")
	}
}

func TestInvalidPlanFailsClosed(t *testing.T) {
	actions, text := parsePlan("not-json")
	if len(actions) != 0 || !strings.Contains(text, "invalid plan") {
		t.Fatalf("unexpected invalid-plan result: %#v %q", actions, text)
	}
}

func TestPlanAcceptsModelCodeFence(t *testing.T) {
	actions, _ := parsePlan("```json\n[{\"name\":\"open_url\",\"arguments\":{\"url\":\"https://racore.xyz\"}}]\n```")
	if len(actions) != 1 || actions[0].Value != "https://racore.xyz" {
		t.Fatalf("unexpected fenced plan: %#v", actions)
	}
}

func TestPlanAcceptsSingleToolObjectFromSmallModel(t *testing.T) {
	actions, _ := parsePlan("```\n{\"name\":\"search_web\",\"arguments\":{\"query\":\"Racore browser\"}}\n```")
	if len(actions) != 1 || actions[0].Type != "search_web" {
		t.Fatalf("unexpected single-object plan: %#v", actions)
	}
}
