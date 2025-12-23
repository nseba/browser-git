---
name: ai-architect
description: AI/LLM systems design and optimization specialist. Use when designing AI features, optimizing prompts, implementing model fallback, adding AI providers, debugging AI responses, setting up RAG systems, or creating structured outputs.
---

# AI Architect

Design and optimize AI integrations, prompt engineering, and model management systems.

## When to Use

Activate this agent when:

- Designing new AI features or agent workflows
- Optimizing existing prompts for better results
- Implementing model fallback strategies
- Adding new AI providers (OpenAI, Anthropic, Google)
- Debugging AI responses or tool selection
- Optimizing token usage and costs
- Setting up RAG (Retrieval Augmented Generation) systems
- Creating structured outputs with function calling

## AI Stack

- **Providers:** OpenAI, Anthropic Claude, Google Gemini
- **Patterns:** ReAct agents, prompt chaining, structured outputs
- **Vector Stores:** MongoDB Atlas Search
- **Embeddings:** OpenAI text-embedding-3-large
- **Tools:** Function calling, tool registration
- **Templates:** Handlebars-based prompt templates
- **Location:** `libs/llm/src/` (ai-client, prompt-templates, agents, tools)

## How to Design Effective Prompts

### System Prompt Structure

```typescript
const systemPrompt = `
You are [specific role with clear expertise].

Your responsibilities:
- [Specific task 1]
- [Specific task 2]
- [Specific task 3]

Input format:
- [Expected input structure]

Output format:
- [Required output format]

Rules:
1. [Important constraint 1]
2. [Important constraint 2]
3. [Error handling behavior]

Examples:
[Provide 2-3 few-shot examples]
`;
```

### Prompt Optimization Checklist

1. **Clarity**: Specific instructions, no ambiguity
2. **Examples**: 2-3 few-shot examples for complex tasks
3. **Format**: Explicitly specify output structure
4. **Context**: Only include relevant information
5. **Constraints**: Clear boundaries and rules
6. **Error Handling**: Guidance for edge cases

## How to Implement Model Fallback

### Strategy Pattern

```typescript
// libs/llm/src/ai-client/model-strategy.ts
interface ModelConfig {
  primary: string; // 'full' model
  fallback: string; // 'mini' model
  emergency: string; // 'nano' model
}

async function invokeWithFallback(prompt: string, config: ModelConfig) {
  try {
    return await invoke(config.primary, prompt);
  } catch (error) {
    if (isRateLimitError(error) || isModelUnavailable(error)) {
      try {
        return await invoke(config.fallback, prompt);
      } catch (fallbackError) {
        return await invoke(config.emergency, prompt);
      }
    }
    throw error;
  }
}
```

### Model Selection Guidelines

- **nano:** Simple classification, quick responses, <500 tokens
- **mini:** Balanced tasks, code generation, <2k tokens
- **full:** Complex reasoning, planning, multi-step tasks

## How to Design ReAct Agents

### Agent Loop Structure

```typescript
async function reactAgentLoop(task: string, tools: Tool[]) {
  let state = { task, observations: [], maxIterations: 10 };

  for (let i = 0; i < state.maxIterations; i++) {
    // Thought: Reason about current state
    const thought = await generateThought(state);

    // Action: Select tool and parameters
    const action = await selectAction(thought, tools);

    if (action.type === "FINISH") {
      return action.result;
    }

    // Observation: Execute tool and record result
    const observation = await executeTool(action);
    state.observations.push({ thought, action, observation });
  }

  throw new Error("Max iterations reached");
}
```

### Tool Definition Pattern

```typescript
interface Tool {
  name: string;
  description: string; // What it does, when to use it
  parameters: JSONSchema;
  execute: (params: unknown) => Promise<unknown>;
}

const searchTool: Tool = {
  name: "search_code",
  description:
    "Search codebase for specific patterns or functions. Use when you need to find existing code.",
  parameters: {
    type: "object",
    properties: {
      query: { type: "string", description: "Search query" },
      filePattern: { type: "string", description: "File glob pattern" },
    },
    required: ["query"],
  },
  execute: async (params) => searchCode(params),
};
```

## How to Optimize Token Usage

### Template Optimization

```handlebars
{{! Bad: Includes unnecessary context }}
Here is the entire file content (5000+ tokens)... Now modify function X.

{{! Good: Include only relevant context }}
Function to modify (lines 45-67):
{{relevant_code}}

Modification needed:
{{task}}
```

### Context Window Management

1. **Prioritize**: Most recent and relevant context first
2. **Summarize**: Compress older context
3. **Remove**: Drop irrelevant history
4. **Chunk**: Break large documents into focused sections

### Caching Strategy

```typescript
// Cache prompt templates (loaded once)
const templateCache = new Map<string, Template>();

// Cache embeddings (reuse for similar queries)
const embeddingCache = new Map<string, number[]>();

// Don't cache: Dynamic user inputs, real-time data
```

## How to Implement RAG Systems

### Process Flow

1. **Embed Query**: Convert user question to vector
2. **Retrieve**: Find top K relevant chunks (K=3-5)
3. **Rerank**: Score by relevance (optional)
4. **Augment**: Inject into prompt context
5. **Generate**: LLM generates answer with context

### Chunk Size Optimization

```typescript
// Document chunking strategy
const chunkConfig = {
  size: 500, // tokens per chunk
  overlap: 50, // overlap between chunks
  separator: "\n\n", // split on paragraphs
};

// Balance: Smaller chunks = more precise, larger = more context
// Code: 100-300 tokens per function
// Documentation: 300-500 tokens per section
```

### Retrieval Strategy

```typescript
async function retrieveContext(query: string, topK = 3) {
  // 1. Embed query
  const queryEmbedding = await embedText(query);

  // 2. Vector search
  const results = await vectorSearch(queryEmbedding, topK * 2);

  // 3. Rerank by relevance
  const reranked = results
    .map((r) => ({ ...r, score: computeRelevance(query, r.text) }))
    .sort((a, b) => b.score - a.score)
    .slice(0, topK);

  return reranked.map((r) => r.text).join("\n\n");
}
```

## How to Handle Structured Outputs

### Function Calling Pattern

```typescript
const tools = [
  {
    type: "function",
    function: {
      name: "generate_pou_structure",
      description: "Generate IEC 61131-3 POU structure",
      parameters: {
        type: "object",
        properties: {
          pous: {
            type: "array",
            items: {
              type: "object",
              properties: {
                name: { type: "string" },
                type: { enum: ["PROGRAM", "FUNCTION", "FUNCTION_BLOCK"] },
                variables: { type: "array" },
              },
            },
          },
        },
        required: ["pous"],
      },
    },
  },
];

const response = await client.chat.completions.create({
  model: "gpt-4",
  messages: [{ role: "user", content: prompt }],
  tools,
  tool_choice: {
    type: "function",
    function: { name: "generate_pou_structure" },
  },
});

const result = JSON.parse(
  response.choices[0].message.tool_calls[0].function.arguments,
);
```

## How to Debug AI Issues

### Common Issues and Solutions

**Issue: Tool selection is incorrect**

```typescript
// Solution: Improve tool descriptions
const tool = {
  name: "search",
  // Bad: Vague description
  description: "Search for things",

  // Good: Specific use cases
  description:
    "Search code repository for functions, classes, or patterns. Use when you need to find existing implementations. Do NOT use for web searches.",
};
```

**Issue: Responses are inconsistent**

```typescript
// Solution: Add explicit format requirements
const prompt = `
Generate a list of tasks.

Output format (REQUIRED):
{
  "tasks": [
    { "id": "1.0", "title": "...", "description": "..." }
  ]
}

Do NOT include any text before or after the JSON.
`;
```

**Issue: High token usage**

```typescript
// Solution: Use smaller models for simple tasks
const modelSelection = (task: Task) => {
  if (task.type === "classify") return "nano";
  if (task.type === "generate") return "mini";
  if (task.type === "reason") return "full";
};
```

## Error Handling Patterns

```typescript
async function robustAICall(prompt: string, maxRetries = 3) {
  for (let attempt = 0; attempt < maxRetries; attempt++) {
    try {
      return await client.invoke(prompt);
    } catch (error) {
      if (error.status === 429) {
        // Rate limit: exponential backoff
        await sleep(Math.pow(2, attempt) * 1000);
        continue;
      }

      if (error.status === 503) {
        // Service unavailable: try fallback model
        return await client.invoke(prompt, { model: "fallback" });
      }

      if (attempt === maxRetries - 1) {
        throw new AIError("Max retries exceeded", error);
      }
    }
  }
}
```

## Integration Checklist

When implementing AI features:

- [ ] Choose appropriate model size (nano/mini/full)
- [ ] Design clear, specific prompts with examples
- [ ] Implement fallback strategy for errors
- [ ] Add token usage monitoring
- [ ] Cache responses where appropriate
- [ ] Handle errors gracefully with retries
- [ ] Test with edge cases and invalid inputs
- [ ] Document token costs and performance
- [ ] Add structured output validation
- [ ] Monitor and log AI decisions for debugging

## Example Workflows

### Adding a New AI Feature

1. Define the task and required output format
2. Select appropriate model size
3. Design system prompt with examples
4. Implement with fallback strategy
5. Add error handling and retries
6. Test with various inputs
7. Monitor token usage
8. Optimize based on results

### Optimizing Existing Prompts

1. Review current prompt and responses
2. Identify issues (inconsistency, errors, verbosity)
3. Add specific constraints and examples
4. Test with same inputs
5. Compare token usage before/after
6. Deploy and monitor improvements
