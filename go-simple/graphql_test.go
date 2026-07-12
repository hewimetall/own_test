package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func sampleJobs() []Job {
	return []Job{
		{
			ID:             1,
			Position:       "Software Engineer",
			Company:        "Apple",
			Description:    "Build products",
			SkillsRequired: []string{"Go", "GraphQL"},
			Location:       "Remote",
			EmploymentType: "full-time",
		},
		{
			ID:             2,
			Position:       "Engineering Manager",
			Company:        "Google",
			Description:    "Lead teams",
			SkillsRequired: []string{"Scrum", "JIRA"},
			Location:       "Mountain View",
			EmploymentType: "full-time",
		},
	}
}

func sampleLoader() ([]Job, error) {
	return sampleJobs(), nil
}

func decodeGraphQLResult(t *testing.T, body string) map[string]any {
	t.Helper()

	var payload map[string]any
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		t.Fatalf("expected valid JSON response, got error %v and body %q", err, body)
	}
	return payload
}

func TestProcessQueryReturnsAllJobs(t *testing.T) {
	result, err := processQuery(`{ jobs { id position skillsRequired } }`, sampleLoader)
	if err != nil {
		t.Fatalf("processQuery returned error: %v", err)
	}

	payload := decodeGraphQLResult(t, result)
	data := payload["data"].(map[string]any)
	jobs := data["jobs"].([]any)
	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs))
	}

	first := jobs[0].(map[string]any)
	if first["position"] != "Software Engineer" {
		t.Fatalf("expected first job position, got %#v", first["position"])
	}
	skills := first["skillsRequired"].([]any)
	if skills[0] != "Go" || skills[1] != "GraphQL" {
		t.Fatalf("expected first job skills, got %#v", skills)
	}
}

func TestProcessQueryReturnsSingleJobByID(t *testing.T) {
	result, err := processQuery(`{ job(id: 2) { id company employmentType } }`, sampleLoader)
	if err != nil {
		t.Fatalf("processQuery returned error: %v", err)
	}

	payload := decodeGraphQLResult(t, result)
	data := payload["data"].(map[string]any)
	job := data["job"].(map[string]any)
	if job["company"] != "Google" || job["employmentType"] != "full-time" {
		t.Fatalf("expected Google full-time job, got %#v", job)
	}
}

func TestProcessQueryReturnsNullForUnknownJob(t *testing.T) {
	result, err := processQuery(`{ job(id: 99) { id company } }`, sampleLoader)
	if err != nil {
		t.Fatalf("processQuery returned error: %v", err)
	}

	payload := decodeGraphQLResult(t, result)
	data := payload["data"].(map[string]any)
	if data["job"] != nil {
		t.Fatalf("expected unknown job to be null, got %#v", data["job"])
	}
}

func TestProcessQueryReturnsNullWhenJobIDIsOmitted(t *testing.T) {
	result, err := processQuery(`{ job { id company } }`, sampleLoader)
	if err != nil {
		t.Fatalf("processQuery returned error: %v", err)
	}

	payload := decodeGraphQLResult(t, result)
	data := payload["data"].(map[string]any)
	if data["job"] != nil {
		t.Fatalf("expected job without id to be null, got %#v", data["job"])
	}
}

func TestProcessQueryReportsGraphQLErrors(t *testing.T) {
	result, err := processQuery(`{ missingField }`, sampleLoader)
	if err != nil {
		t.Fatalf("processQuery returned Go error for GraphQL error: %v", err)
	}

	payload := decodeGraphQLResult(t, result)
	errorsValue, ok := payload["errors"].([]any)
	if !ok || len(errorsValue) == 0 {
		t.Fatalf("expected GraphQL errors, got %#v", payload)
	}
}

func TestProcessQueryRequiresLoader(t *testing.T) {
	if _, err := processQuery(`{ jobs { id } }`, nil); err == nil {
		t.Fatal("expected nil loader to return an error")
	}
}

func TestProcessQueryIncludesLoaderErrors(t *testing.T) {
	result, err := processQuery(`{ jobs { id } }`, func() ([]Job, error) {
		return nil, errors.New("loader failed")
	})
	if err != nil {
		t.Fatalf("processQuery returned Go error for resolver error: %v", err)
	}
	if !strings.Contains(result, "loader failed") {
		t.Fatalf("expected loader error in GraphQL response, got %s", result)
	}
}

func TestProcessQueryIncludesSingleJobLoaderErrors(t *testing.T) {
	result, err := processQuery(`{ job(id: 1) { id } }`, func() ([]Job, error) {
		return nil, errors.New("single job loader failed")
	})
	if err != nil {
		t.Fatalf("processQuery returned Go error for resolver error: %v", err)
	}
	if !strings.Contains(result, "single job loader failed") {
		t.Fatalf("expected single-job loader error in GraphQL response, got %s", result)
	}
}

func TestGQLHandlerWritesJSONResponse(t *testing.T) {
	handler := gqlHandlerWithLoader(sampleLoader)
	request := httptest.NewRequest(http.MethodPost, "/graphql", strings.NewReader(`{"query":"{ jobs { id } }"}`))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d and body %q", recorder.Code, recorder.Body.String())
	}
	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected JSON content type, got %q", contentType)
	}
	payload := decodeGraphQLResult(t, recorder.Body.String())
	if payload["data"] == nil {
		t.Fatalf("expected data in response, got %#v", payload)
	}
}

func TestGQLHandlerRejectsMissingBody(t *testing.T) {
	recorder := httptest.NewRecorder()
	gqlHandlerWithLoader(sampleLoader).ServeHTTP(recorder, &http.Request{Method: http.MethodPost})

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for missing body, got %d", recorder.Code)
	}
}

func TestGQLHandlerRejectsInvalidJSON(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/graphql", strings.NewReader(`not-json`))
	recorder := httptest.NewRecorder()

	gqlHandlerWithLoader(sampleLoader).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for invalid JSON, got %d", recorder.Code)
	}
}

func TestGQLHandlerReturnsServerErrorForInvalidLoader(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/graphql", strings.NewReader(`{"query":"{ jobs { id } }"}`))
	recorder := httptest.NewRecorder()

	gqlHandlerWithLoader(nil).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500 for invalid loader, got %d", recorder.Code)
	}
}

func TestRetrieveJobsFromFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "jobs.json")
	data, err := json.Marshal(sampleJobs())
	if err != nil {
		t.Fatalf("marshal sample jobs: %v", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write test jobs file: %v", err)
	}

	jobs, err := retrieveJobsFromFile(path)()
	if err != nil {
		t.Fatalf("retrieveJobsFromFile returned error: %v", err)
	}
	if len(jobs) != 2 || jobs[1].Company != "Google" {
		t.Fatalf("expected decoded jobs, got %#v", jobs)
	}
}

func TestRetrieveJobsFromFileReturnsOpenError(t *testing.T) {
	_, err := retrieveJobsFromFile(filepath.Join(t.TempDir(), "missing.json"))()
	if err == nil {
		t.Fatal("expected missing file error")
	}
}

func TestRetrieveJobsFromFileReturnsParseError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "jobs.json")
	if err := os.WriteFile(path, []byte(`not-json`), 0o600); err != nil {
		t.Fatalf("write invalid jobs file: %v", err)
	}

	_, err := retrieveJobsFromFile(path)()
	if err == nil {
		t.Fatal("expected invalid JSON error")
	}
}
