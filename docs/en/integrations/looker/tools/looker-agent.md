---
title: "looker-agent"
type: docs
weight: 1
description: >
  Manage Looker Agents
aliases:
- /resources/tools/looker-agent
---

## About

The `looker-agent` tool allows LLMs to manage Looker Agents. It supports listing, retrieving, creating, updating, and deleting agents using the Looker Go SDK.

## Compatible Sources

{{< compatible-sources >}}

## Configuration

To use the `looker-agent` tool, you must define it in your `server.yaml` file.

```yaml
kind: tools
name: agent_manage
type: looker-agent
source: my_looker_source
description: |
  Manage Looker AI Agents. This tool allows you to perform various operations on Looker Agents, 
  including listing all available agents, retrieving details for a specific agent, 
  creating a new agent, updating an existing one, and deleting an agent.

  Parameters:
  - operation (required): The action to perform.
    - 'list': Returns a list of all existing agents.
    - 'get': Retrieves detailed information about a specific agent. Requires 'agent_id'.
    - 'create': Creates a new Looker AI Agent. Requires 'name'. Optional 'instructions', 'sources', 'code_interpreter'.
    - 'update': Updates an existing Looker AI Agent. Requires 'agent_id'. Optional 'name', 'instructions', 'sources', 'code_interpreter'.
    - 'delete': Removes an existing agent. Requires 'agent_id'.
  - agent_id (optional): The unique identifier of the agent. Required for 'get', 'update', and 'delete' operations.
  - name (optional): The display name for the agent. Required for 'create' operation.
  - instructions (optional): The system prompt or instructions for the agent. Used for 'create' and 'update' operations.
  - sources (optional): A list of JSON-encoded data sources for the agent. Each source should be a JSON string with 'model' and 'explore' keys.
  - code_interpreter (optional): A boolean value to enable or disable Code Interpreter for this Agent.
```

## Parameters

| **Parameter** | **Type** | **Required** | **Description** |
|:-------------|:--------:|:------------:|:----------------|
| `operation` | `string` | Yes | The operation to perform. Must be one of: `list`, `get`, `create`, `update`, or `delete`. |
| `agent_id` | `string` | No | The ID of the agent. Required for `get`, `update`, and `delete` operations. |
| `name` | `string` | No | The name of the agent. Required for `create` operation. |
| `instructions` | `string` | No | The instructions (system prompt) for the agent. Used for `create` and `update` operations. |
| sources | array | No | Optional. A list of JSON-encoded data sources, where each is a string with 'model' and 'explore' keys. |
| `code_interpreter` | `boolean` | No | Optional. Enables Code Interpreter for this Agent. |

## Operations

### List Agents
Retrieve a list of all agents.
```json
{
  "operation": "list"
}
```

### Get Agent
Retrieve details of a specific agent by its ID.
```json
{
  "operation": "get",
  "agent_id": "12345"
}
```

### Create Agent
Create a new agent with a name, instructions, data sources, and Code Interpreter enabled.
```json
{
  "operation": "create",
  "name": "My AI Assistant",
  "instructions": "You are a helpful data analyst. Always provide clear summaries.",
  "sources": [
    "{\"model\": \"thelook\", \"explore\": \"orders\"}"
  ],
  "code_interpreter": true
}
```

### Update Agent
Update an existing agent's instructions and disable Code Interpreter.
```json
{
  "operation": "update",
  "agent_id": "12345",
  "instructions": "New updated instructions for the agent.",
  "code_interpreter": false
}
```

### Delete Agent
Delete an agent by its ID.
```json
{
  "operation": "delete",
  "agent_id": "12345"
}
```

## Dependencies
This tool requires the underlying Looker Go SDK to support the `Agent` API resource (v0.26.6+).
