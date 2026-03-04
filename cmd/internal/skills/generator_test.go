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

package skills

import (
	"strings"
	"testing"

	"github.com/googleapis/genai-toolbox/internal/server"
	"github.com/googleapis/genai-toolbox/internal/tools"
	"github.com/googleapis/genai-toolbox/internal/util/parameters"
)

func TestGenerateReferenceMarkdown(t *testing.T) {
	tests := []struct {
		name         string
		toolName     string
		params       []parameters.ParameterManifest
		wantContains []string
		wantErr      bool
	}{
		{
			name:         "empty parameters",
			toolName:     "tool1",
			params:       []parameters.ParameterManifest{},
			wantContains: []string{"# tool1", "This tool has no parameters."},
		},
		{
			name:     "single required string parameter",
			toolName: "tool1",
			params: []parameters.ParameterManifest{
				{
					Name:        "param1",
					Description: "A test parameter",
					Type:        "string",
					Required:    true,
				},
			},
			wantContains: []string{
				"# tool1",
				"## Parameters",
				"```json",
				`"type": "object"`,
				`"properties": {`,
				`"param1": {`,
				`"type": "string"`,
				`"description": "A test parameter"`,
				`"required": [`,
				`"param1"`,
			},
		},
		{
			name:     "mixed parameters with defaults",
			toolName: "tool1",
			params: []parameters.ParameterManifest{
				{
					Name:        "param1",
					Description: "Param 1",
					Type:        "string",
					Required:    true,
				},
				{
					Name:        "param2",
					Description: "Param 2",
					Type:        "integer",
					Default:     42,
					Required:    false,
				},
			},
			wantContains: []string{
				`"param1": {`,
				`"param2": {`,
				`"default": 42`,
				`"required": [`,
				`"param1"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateReferenceMarkdown(tt.toolName, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateReferenceMarkdown() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("generateReferenceMarkdown() result missing expected string: %s\nGot:\n%s", want, got)
				}
			}
		})
	}
}

func TestGenerateSkillMarkdown(t *testing.T) {
	toolsMap := map[string]tools.Tool{
		"tool1": server.MockTool{
			Description: "First tool",
			Params: []parameters.Parameter{
				parameters.NewStringParameter("p1", "d1"),
			},
		},
		"tool2": server.MockTool{
			Description: "Second tool",
			Params:      []parameters.Parameter{},
		},
	}

	got, err := generateSkillMarkdown("MySkill", "My Description", toolsMap)
	if err != nil {
		t.Fatalf("generateSkillMarkdown() error = %v", err)
	}

	expectedSubstrings := []string{
		"name: MySkill",
		"description: My Description",
		"## Usage",
		"All scripts can be executed using Node.js",
		"**Bash:**",
		"`node <skill_dir>/scripts/<script_name>.js '{\"<param_name>\": \"<param_value>\"}'`",
		"**PowerShell:**",
		"`node <skill_dir>/scripts/<script_name>.js '{\"<param_name>\": \"<param_value>\"}'`",
		"## Scripts",
		"### tool1",
		"First tool",
		"[Parameter Reference](<skill_dir>/references/tool1.md)",
		"### tool2",
		"Second tool",
		"This tool has no parameters.",
	}

	for _, s := range expectedSubstrings {
		if !strings.Contains(got, s) {
			t.Errorf("generateSkillMarkdown() missing substring %q\nGot:\n%s", s, got)
		}
	}
}

func TestGenerateScriptContent(t *testing.T) {
	tests := []struct {
		name         string
		toolName     string
		configArgs   string
		wantContains []string
	}{
		{
			name:       "basic script",
			toolName:   "test-tool",
			configArgs: `"--prebuilt", "test"`,
			wantContains: []string{
				"#!/usr/bin/env node",
				"// Copyright 2026 Google LLC",
				"// Licensed under the Apache License, Version 2.0 (the \"License\");",
				`const toolName = "test-tool";`,
				`const configArgs = ["--prebuilt", "test"];`,
				`const toolboxArgs = [...configArgs, "invoke", toolName, ...args];`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateScriptContent(tt.toolName, tt.configArgs)
			if err != nil {
				t.Fatalf("generateScriptContent() error = %v", err)
			}

			for _, s := range tt.wantContains {
				if !strings.Contains(got, s) {
					t.Errorf("generateScriptContent() missing substring %q\nGot:\n%s", s, got)
				}
			}
		})
	}
}
