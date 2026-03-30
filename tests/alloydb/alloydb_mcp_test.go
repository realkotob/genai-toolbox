// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package alloydb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/googleapis/genai-toolbox/internal/testutils"
	"github.com/googleapis/genai-toolbox/tests"
)

func TestAlloyDBListTools(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	toolsFile := getAlloyDBToolsConfig()

	// Start the toolbox server
	cmd, cleanup, err := tests.StartCmd(ctx, toolsFile)
	if err != nil {
		t.Fatalf("command initialization returned an error: %v", err)
	}
	defer cleanup()

	waitCtx, cancelWait := context.WithTimeout(ctx, 20*time.Second)
	defer cancelWait()
	out, err := testutils.WaitForString(waitCtx, regexp.MustCompile(`Server ready to serve`), cmd.Out)
	if err != nil {
		t.Logf("toolbox command logs: \n%s", out)
		t.Fatalf("toolbox didn't start successfully: %v", err)
	}

	// Verify list of tools
	expectedTools := []tests.MCPToolManifest{
		{
			Name:        "alloydb-list-clusters",
			Description: "Lists all AlloyDB clusters in a given project and location.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"project": map[string]any{
						"description": "The GCP project ID to list clusters for.",
						"type":        "string",
					},
					"location": map[string]any{
						"default":     "-",
						"description": "Optional: The location to list clusters in (e.g., 'us-central1'). Use '-' to list clusters across all locations.(Default: '-')",
						"type":        "string",
					},
				},
				"required": []any{"project"},
			},
		},
		{
			Name:        "alloydb-list-users",
			Description: "Lists all AlloyDB users within a specific cluster.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"cluster": map[string]any{
						"description": "The ID of the cluster to list users from.",
						"type":        "string",
					},
					"location": map[string]any{
						"description": "The location of the cluster (e.g., 'us-central1').",
						"type":        "string",
					},
					"project": map[string]any{
						"description": "The GCP project ID.",
						"type":        "string",
					},
				},
				"required": []any{"project", "location", "cluster"},
			},
		},
	}

	tests.RunMCPToolsListMethod(t, expectedTools)
}

func TestAlloyDBCallTool(t *testing.T) {
	vars := getAlloyDBVars(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	toolsFile := getAlloyDBToolsConfig()

	cmd, cleanup, err := tests.StartCmd(ctx, toolsFile)
	if err != nil {
		t.Fatalf("command initialization returned an error: %v", err)
	}
	defer cleanup()

	waitCtx, cancelWait := context.WithTimeout(ctx, 20*time.Second)
	defer cancelWait()
	out, err := testutils.WaitForString(waitCtx, regexp.MustCompile(`Server ready to serve`), cmd.Out)
	if err != nil {
		t.Logf("toolbox command logs: \n%s", out)
		t.Fatalf("toolbox didn't start successfully: %v", err)
	}

	// Test calling "alloydb-list-clusters"
	args := map[string]any{
		"project":  vars["project"],
		"location": vars["location"],
	}

	wantContains := fmt.Sprintf(`"name":"projects/%s/locations/%s/clusters/%s"`, vars["project"], vars["location"], vars["cluster"])

	tests.RunMCPCustomToolCallMethod(t, "alloydb-list-clusters", args, wantContains)

	// Negative cases from legacy runAlloyDBMCPToolCallMethod
	t.Run("MCP Invoke my-fail-tool missing project", func(t *testing.T) {
		statusCode, mcpResp, err := tests.InvokeMCPTool(t, "my-fail-tool", map[string]any{"location": vars["location"]}, nil)
		if err != nil {
			t.Fatalf("native error executing %s: %s", "my-fail-tool", err)
		}
		if statusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", statusCode)
		}
		tests.AssertMCPError(t, mcpResp, `parameter "project" is required`)
	})

	t.Run("MCP Invoke invalid tool", func(t *testing.T) {
		statusCode, mcpResp, err := tests.InvokeMCPTool(t, "non-existent-tool", map[string]any{}, nil)
		if err != nil {
			t.Fatalf("native error executing %s: %s", "non-existent-tool", err)
		}
		if statusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", statusCode)
		}
		tests.AssertMCPError(t, mcpResp, `tool with name "non-existent-tool" does not exist`)
	})

	t.Run("MCP Invoke tool without required parameters", func(t *testing.T) {
		statusCode, mcpResp, err := tests.InvokeMCPTool(t, "alloydb-list-clusters", map[string]any{"location": vars["location"]}, nil)
		if err != nil {
			t.Fatalf("native error executing %s: %s", "alloydb-list-clusters", err)
		}
		if statusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", statusCode)
		}
		tests.AssertMCPError(t, mcpResp, `parameter "project" is required`)
	})
}

func TestAlloyDBListClusters(t *testing.T) {
	vars := getAlloyDBVars(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	toolsFile := getAlloyDBToolsConfig()

	cmd, cleanup, err := tests.StartCmd(ctx, toolsFile)
	if err != nil {
		t.Fatalf("command initialization returned an error: %v", err)
	}
	defer cleanup()

	waitCtx, cancelWait := context.WithTimeout(ctx, 20*time.Second)
	defer cancelWait()
	out, err := testutils.WaitForString(waitCtx, regexp.MustCompile(`Server ready to serve`), cmd.Out)
	if err != nil {
		t.Logf("toolbox command logs: \n%s", out)
		t.Fatalf("toolbox didn't start successfully: %v", err)
	}

	wantForAllLocations := []string{
		fmt.Sprintf("projects/%s/locations/us-central1/clusters/alloydb-ai-nl-testing", vars["project"]),
		fmt.Sprintf("projects/%s/locations/us-central1/clusters/alloydb-pg-testing", vars["project"]),
	}

	t.Run("list clusters for all locations", func(t *testing.T) {
		args := map[string]any{"project": vars["project"], "location": "-"}
		statusCode, mcpResp, err := tests.InvokeMCPTool(t, "alloydb-list-clusters", args, nil)
		if err != nil {
			t.Fatalf("native error executing: %s", err)
		}
		if statusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", statusCode)
		}
		if mcpResp.Result.IsError {
			t.Fatalf("returned error result: %v", mcpResp.Result)
		}

		got := mcpResp.Result.Content[0].Text
		for _, want := range wantForAllLocations {
			if !regexp.MustCompile(want).MatchString(got) {
				t.Errorf("Expected substring not found: %q", want)
			}
		}
	})

	t.Run("list clusters missing project", func(t *testing.T) {
		args := map[string]any{"location": vars["location"]}
		statusCode, mcpResp, err := tests.InvokeMCPTool(t, "alloydb-list-clusters", args, nil)
		if err != nil {
			t.Fatalf("native error executing: %s", err)
		}
		if statusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", statusCode)
		}
		tests.AssertMCPError(t, mcpResp, `parameter "project" is required`)
	})
}

type mockAlloyDBTransportMCP struct {
	transport http.RoundTripper
	url       *url.URL
}

func (t *mockAlloyDBTransportMCP) RoundTrip(req *http.Request) (*http.Response, error) {
	if strings.HasPrefix(req.URL.String(), "https://alloydb.googleapis.com") {
		req.URL.Scheme = t.url.Scheme
		req.URL.Host = t.url.Host
	}
	return t.transport.RoundTrip(req)
}

type mockAlloyDBHandlerMCP struct {
	t       *testing.T
	idParam string
}

func (h *mockAlloyDBHandlerMCP) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.Contains(r.UserAgent(), "genai-toolbox/") {
		h.t.Errorf("User-Agent header not found")
	}

	id := r.URL.Query().Get(h.idParam)

	var response string
	var statusCode int

	switch id {
	case "c1-success":
		response = `{
			"name": "projects/p1/locations/l1/operations/mock-operation-success",
			"metadata": {
				"verb": "create",
				"target": "projects/p1/locations/l1/clusters/c1-success"
			}
		}`
		statusCode = http.StatusOK
	case "c2-api-failure":
		response = `{"error":{"message":"internal api error"}}`
		statusCode = http.StatusInternalServerError
	case "i1-success":
		response = `{
			"metadata": {
				"@type": "type.googleapis.com/google.cloud.alloydb.v1.OperationMetadata",
				"target": "projects/p1/locations/l1/clusters/c1/instances/i1-success",
				"verb": "create",
				"requestedCancellation": false,
				"apiVersion": "v1"
			},
			"name": "projects/p1/locations/l1/operations/mock-operation-success"
		}`
		statusCode = http.StatusOK
	case "i2-api-failure":
		response = `{"error":{"message":"internal api error"}}`
		statusCode = http.StatusInternalServerError
	case "u1-iam-success":
		response = `{
			"databaseRoles": ["alloydbiamuser"],
			"name": "projects/p1/locations/l1/clusters/c1/users/u1-iam-success",
			"userType": "ALLOYDB_IAM_USER"
		}`
		statusCode = http.StatusOK
	case "u2-builtin-success":
		response = `{
			"databaseRoles": ["alloydbsuperuser"],
			"name": "projects/p1/locations/l1/clusters/c1/users/u2-builtin-success",
			"userType": "ALLOYDB_BUILT_IN"
		}`
		statusCode = http.StatusOK
	case "u3-api-failure":
		response = `{"error":{"message":"user internal api error"}}`
		statusCode = http.StatusInternalServerError
	default:
		http.Error(w, fmt.Sprintf("unhandled %s in mock server: %s", h.idParam, id), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if _, err := w.Write([]byte(response)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func setupTestServerMCP(t *testing.T, idParam string) func() {
	handler := &mockAlloyDBHandlerMCP{t: t, idParam: idParam}
	server := httptest.NewServer(handler)

	serverURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("failed to parse server URL: %v", err)
	}

	originalTransport := http.DefaultClient.Transport
	if originalTransport == nil {
		originalTransport = http.DefaultTransport
	}
	http.DefaultClient.Transport = &mockAlloyDBTransportMCP{
		transport: originalTransport,
		url:       serverURL,
	}

	return func() {
		server.Close()
		http.DefaultClient.Transport = originalTransport
	}
}

func TestAlloyDBCreateClusterMCP(t *testing.T) {
	cleanup := setupTestServerMCP(t, "clusterId")
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	args := []string{"--enable-api"}
	toolsFile := getAlloyDBToolsConfig()
	cmd, cleanupCmd, err := tests.StartCmd(ctx, toolsFile, args...)
	if err != nil {
		t.Fatalf("command initialization returned an error: %v", err)
	}
	defer cleanupCmd()

	waitCtx, cancelWait := context.WithTimeout(ctx, 10*time.Second)
	defer cancelWait()
	out, err := testutils.WaitForString(waitCtx, regexp.MustCompile(`Server ready to serve`), cmd.Out)
	if err != nil {
		t.Logf("toolbox command logs: \n%s", out)
		t.Fatalf("toolbox didn't start successfully: %s", err)
	}

	tcs := []struct {
		name           string
		body           string
		want           string
		wantStatusCode int
	}{
		{
			name:           "successful creation",
			body:           `{"project": "p1", "location": "l1", "cluster": "c1-success", "password": "p1"}`,
			want:           `{"name":"projects/p1/locations/l1/operations/mock-operation-success", "metadata": {"verb": "create", "target": "projects/p1/locations/l1/clusters/c1-success"}}`,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "api failure",
			body:           `{"project": "p1", "location": "l1", "cluster": "c2-api-failure", "password": "p1"}`,
			want:           `{"error":"error processing GCP request: error creating AlloyDB cluster: googleapi: Error 500: internal api error"}`,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "missing project",
			body:           `{"location": "l1", "cluster": "c1", "password": "p1"}`,
			want:           `{"error":"parameter \"project\" is required"}`,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "missing cluster",
			body:           `{"project": "p1", "location": "l1", "password": "p1"}`,
			want:           `{"error":"parameter \"cluster\" is required"}`,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "missing password",
			body:           `{"project": "p1", "location": "l1", "cluster": "c1"}`,
			want:           `{"error":"parameter \"password\" is required"}`,
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			var args map[string]any
			if err := json.Unmarshal([]byte(tc.body), &args); err != nil {
				t.Fatalf("failed to unmarshal body: %v", err)
			}

			statusCode, mcpResp, err := tests.InvokeMCPTool(t, "alloydb-create-cluster", args, nil)
			if err != nil {
				t.Fatalf("native error executing %s: %s", "alloydb-create-cluster", err)
			}

			if statusCode != http.StatusOK {
				t.Fatalf("expected status 200, got %d", statusCode)
			}

			if tc.name == "successful creation" {
				if mcpResp.Result.IsError {
					t.Fatalf("expected success, got error result: %v", mcpResp.Result)
				}
				gotStr := mcpResp.Result.Content[0].Text
				var got, want map[string]any
				if err := json.Unmarshal([]byte(gotStr), &got); err != nil {
					t.Fatalf("failed to unmarshal result: %v", err)
				}
				if err := json.Unmarshal([]byte(tc.want), &want); err != nil {
					t.Fatalf("failed to unmarshal want: %v", err)
				}
				if diff := cmp.Diff(want, got); diff != "" {
					t.Errorf("unexpected result (-want +got):\n%s", diff)
				}
			} else {
				var wantMap map[string]string
				if err := json.Unmarshal([]byte(tc.want), &wantMap); err != nil {
					t.Fatalf("failed to unmarshal want: %v", err)
				}
				tests.AssertMCPError(t, mcpResp, wantMap["error"])
			}
		})
	}
}

func TestAlloyDBCreateInstanceMCP(t *testing.T) {
	cleanup := setupTestServerMCP(t, "instanceId")
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	args := []string{"--enable-api"}
	toolsFile := getAlloyDBToolsConfig()
	cmd, cleanupCmd, err := tests.StartCmd(ctx, toolsFile, args...)
	if err != nil {
		t.Fatalf("command initialization returned an error: %v", err)
	}
	defer cleanupCmd()

	waitCtx, cancelWait := context.WithTimeout(ctx, 10*time.Second)
	defer cancelWait()
	out, err := testutils.WaitForString(waitCtx, regexp.MustCompile(`Server ready to serve`), cmd.Out)
	if err != nil {
		t.Logf("toolbox command logs: \n%s", out)
		t.Fatalf("toolbox didn't start successfully: %s", err)
	}

	tcs := []struct {
		name           string
		body           string
		want           string
		wantStatusCode int
	}{
		{
			name:           "successful creation",
			body:           `{"project": "p1", "location": "l1", "cluster": "c1", "instance": "i1-success", "instanceType": "PRIMARY", "displayName": "i1-success"}`,
			want:           `{"metadata":{"@type":"type.googleapis.com/google.cloud.alloydb.v1.OperationMetadata","target":"projects/p1/locations/l1/clusters/c1/instances/i1-success","verb":"create","requestedCancellation":false,"apiVersion":"v1"},"name":"projects/p1/locations/l1/operations/mock-operation-success"}`,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "api failure",
			body:           `{"project": "p1", "location": "l1", "cluster": "c1", "instance": "i2-api-failure", "instanceType": "PRIMARY", "displayName": "i1-success"}`,
			want:           `{"error":"error processing GCP request: error creating AlloyDB instance: googleapi: Error 500: internal api error"}`,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "missing project",
			body:           `{"location": "l1", "cluster": "c1", "instance": "i1", "instanceType": "PRIMARY"}`,
			want:           `{"error":"parameter \"project\" is required"}`,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "missing cluster",
			body:           `{"project": "p1", "location": "l1", "instance": "i1", "instanceType": "PRIMARY"}`,
			want:           `{"error":"parameter \"cluster\" is required"}`,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "missing location",
			body:           `{"project": "p1", "cluster": "c1", "instance": "i1", "instanceType": "PRIMARY"}`,
			want:           `{"error":"parameter \"location\" is required"}`,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "missing instance",
			body:           `{"project": "p1", "location": "l1", "cluster": "c1", "instanceType": "PRIMARY"}`,
			want:           `{"error":"parameter \"instance\" is required"}`,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "invalid instanceType",
			body:           `{"project": "p1", "location": "l1", "cluster": "c1", "instance": "i1", "instanceType": "INVALID", "displayName": "invalid"}`,
			want:           `{"error":"invalid 'instanceType' parameter; expected 'PRIMARY' or 'READ_POOL'"}`,
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			var args map[string]any
			if err := json.Unmarshal([]byte(tc.body), &args); err != nil {
				t.Fatalf("failed to unmarshal body: %v", err)
			}

			statusCode, mcpResp, err := tests.InvokeMCPTool(t, "alloydb-create-instance", args, nil)
			if err != nil {
				t.Fatalf("native error executing %s: %s", "alloydb-create-instance", err)
			}

			if statusCode != http.StatusOK {
				t.Fatalf("expected status 200, got %d", statusCode)
			}

			if tc.name == "successful creation" {
				if mcpResp.Result.IsError {
					t.Fatalf("expected success, got error result: %v", mcpResp.Result)
				}
				gotStr := mcpResp.Result.Content[0].Text
				var got, want map[string]any
				if err := json.Unmarshal([]byte(gotStr), &got); err != nil {
					t.Fatalf("failed to unmarshal result: %v", err)
				}
				if err := json.Unmarshal([]byte(tc.want), &want); err != nil {
					t.Fatalf("failed to unmarshal want: %v", err)
				}
				if !reflect.DeepEqual(want, got) {
					t.Errorf("unexpected result:\n- want: %+v\n-  got: %+v", want, got)
				}
			} else {
				var wantMap map[string]string
				if err := json.Unmarshal([]byte(tc.want), &wantMap); err != nil {
					t.Fatalf("failed to unmarshal want: %v", err)
				}
				tests.AssertMCPError(t, mcpResp, wantMap["error"])
			}
		})
	}
}

func TestAlloyDBCreateUserMCP(t *testing.T) {
	cleanup := setupTestServerMCP(t, "userId")
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	args := []string{"--enable-api"}
	toolsFile := getAlloyDBToolsConfig()
	cmd, cleanupCmd, err := tests.StartCmd(ctx, toolsFile, args...)
	if err != nil {
		t.Fatalf("command initialization returned an error: %v", err)
	}
	defer cleanupCmd()

	waitCtx, cancelWait := context.WithTimeout(ctx, 10*time.Second)
	defer cancelWait()
	out, err := testutils.WaitForString(waitCtx, regexp.MustCompile(`Server ready to serve`), cmd.Out)
	if err != nil {
		t.Logf("toolbox command logs: \n%s", out)
		t.Fatalf("toolbox didn't start successfully: %s", err)
	}

	tcs := []struct {
		name           string
		body           string
		want           string
		wantStatusCode int
	}{
		{
			name:           "successful creation IAM user",
			body:           `{"project": "p1", "location": "l1", "cluster": "c1", "user": "u1-iam-success", "userType": "ALLOYDB_IAM_USER"}`,
			want:           `{"databaseRoles": ["alloydbiamuser"], "name": "projects/p1/locations/l1/clusters/c1/users/u1-iam-success", "userType": "ALLOYDB_IAM_USER"}`,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "successful creation builtin user",
			body:           `{"project": "p1", "location": "l1", "cluster": "c1", "user": "u2-builtin-success", "userType": "ALLOYDB_BUILT_IN", "password": "pass123", "databaseRoles": ["alloydbsuperuser"]}`,
			want:           `{"databaseRoles": ["alloydbsuperuser"], "name": "projects/p1/locations/l1/clusters/c1/users/u2-builtin-success", "userType": "ALLOYDB_BUILT_IN"}`,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "api failure",
			body:           `{"project": "p1", "location": "l1", "cluster": "c1", "user": "u3-api-failure", "userType": "ALLOYDB_IAM_USER"}`,
			want:           `{"error":"error processing GCP request: error creating AlloyDB user: googleapi: Error 500: user internal api error"}`,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "missing project",
			body:           `{"location": "l1", "cluster": "c1", "user": "u-fail", "userType": "ALLOYDB_IAM_USER"}`,
			want:           `{"error":"parameter \"project\" is required"}`,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "missing cluster",
			body:           `{"project": "p1", "location": "l1", "user": "u-fail", "userType": "ALLOYDB_IAM_USER"}`,
			want:           `{"error":"parameter \"cluster\" is required"}`,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "missing location",
			body:           `{"project": "p1", "cluster": "c1", "user": "u-fail", "userType": "ALLOYDB_IAM_USER"}`,
			want:           `{"error":"parameter \"location\" is required"}`,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "missing user",
			body:           `{"project": "p1", "location": "l1", "cluster": "c1", "userType": "ALLOYDB_IAM_USER"}`,
			want:           `{"error":"parameter \"user\" is required"}`,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "missing userType",
			body:           `{"project": "p1", "location": "l1", "cluster": "c1", "user": "u-fail"}`,
			want:           `{"error":"parameter \"userType\" is required"}`,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "missing password for builtin user",
			body:           `{"project": "p1", "location": "l1", "cluster": "c1", "user": "u-fail", "userType": "ALLOYDB_BUILT_IN"}`,
			want:           `{"error":"password is required when userType is ALLOYDB_BUILT_IN"}`,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "invalid userType",
			body:           `{"project": "p1", "location": "l1", "cluster": "c1", "user": "u-fail", "userType": "invalid"}`,
			want:           `{"error":"invalid or missing 'userType' parameter; expected 'ALLOYDB_BUILT_IN' or 'ALLOYDB_IAM_USER'"}`,
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			var args map[string]any
			if err := json.Unmarshal([]byte(tc.body), &args); err != nil {
				t.Fatalf("failed to unmarshal body: %v", err)
			}

			statusCode, mcpResp, err := tests.InvokeMCPTool(t, "alloydb-create-user", args, nil)
			if err != nil {
				t.Fatalf("native error executing %s: %s", "alloydb-create-user", err)
			}

			if statusCode != http.StatusOK {
				t.Fatalf("expected status 200, got %d", statusCode)
			}

			if tc.name == "successful creation IAM user" || tc.name == "successful creation builtin user" {
				if mcpResp.Result.IsError {
					t.Fatalf("expected success, got error result: %v", mcpResp.Result)
				}
				gotStr := mcpResp.Result.Content[0].Text
				var got, want map[string]any
				if err := json.Unmarshal([]byte(gotStr), &got); err != nil {
					t.Fatalf("failed to unmarshal result string: %v. Result: %s", err, gotStr)
				}
				if err := json.Unmarshal([]byte(tc.want), &want); err != nil {
					t.Fatalf("failed to unmarshal want string: %v. Want: %s", err, tc.want)
				}
				if diff := cmp.Diff(want, got); diff != "" {
					t.Errorf("unexpected result map (-want +got):\n%s", diff)
				}
			} else {
				var wantMap map[string]string
				if err := json.Unmarshal([]byte(tc.want), &wantMap); err != nil {
					t.Fatalf("failed to unmarshal want: %v", err)
				}
				tests.AssertMCPError(t, mcpResp, wantMap["error"])
			}
		})
	}
}
