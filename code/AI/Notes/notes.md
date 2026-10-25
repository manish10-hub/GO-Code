### Tokenizer

Tokenizer converts text → token IDs.
LLMs operate on token IDs, not raw text.
Modern models primarily use subword tokenization.
BPE, WordPiece, and SentencePiece are the three major algorithms to know.
GPT-family tokenizers are based on byte-level BPE concepts, enabling representation of any input without unknown-character issues.
Pricing, context window, and latency are all fundamentally based on the number of tokens processed.

Why not split by spaces?
Huge vocabulary
Unknown words
Multiple Languages
Emojis
URLs

What is Subword Tokenization?
Small vocabulary
Represent every word
Handle unknown words
Efficient
Language independent


Different Tokenization Algorithms
1: Character Tokenization
Advantages
Tiny vocabulary
No unknown words

Disadvantages
Very long sequences.
Transformers become slow.

2. Word Tokenization
Advantages
Easy.

Disadvantages
Vocabulary may exceed millions.
Unknown words impossible.
Not suitable for LLMs.

3. Subword Tokenization: Most modern LLMs use this.
   Advantages
Small vocabulary
Handles unknown words
Efficient


Popular Algorithms
1. Byte Pair Encoding (BPE)
2. 2. WordPiece
   Used by BERT, DistilBERT

Similar to BPE.
Difference: Instead of frequency,choose merge based on maximum likelihood.
playing: play ##ing

3. SentencePiece
Used by T5,LLaMA

Very popular.
Interesting property: No pre-tokenization.Even spaces become tokens.
Hello World: ▁Hello ▁World

Q1. Why not use word tokenization?

Because vocabulary becomes huge, unknown words cannot be represented, and it doesn't work well across languages.

Q2. Why is BPE better?

It learns common subwords, reduces vocabulary size, and can tokenize unseen words by combining known pieces.

Q3. Why is SentencePiece popular?

It doesn't rely on spaces, works well for multilingual text, and treats whitespace consistently, making it suitable for many modern LLMs.

Q4. Why does "internationalization" consume more tokens than "cat"?

Because tokenizers split long or uncommon words into multiple subword pieces, while frequent short words are often a single token.

Q5. Does the model understand text directly?

No. The pipeline is:

Text
   ↓
Tokenizer
   ↓
Token IDs
   ↓
Embeddings
   ↓
Transformer
   ↓
Output Tokens
   ↓
Detokenizer
   ↓
Text


### Detailed Explanation of AI **API** Integration

Here's a detailed explanation of every topic — what it is, why it matters for AI API integration, and how it works in Go.

1. HTTP Request Construction
What it is
The foundation of everything. Every AI API (OpenAI, Anthropic, Ollama) is just an HTTP server. You send a request, you get a response. No magic.

Why it matters for AI
You are not using a high-level SDK — you are building one. Every call to Claude or OpenAI is ultimately:

```
POST https://api.openai.com/v1/chat/completions
Content-Type: application/json
Authorization: Bearer sk-...

{ "model": "gpt-4o", "messages": [...] }
```

How it works in Go
```
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

func main() {
    // Step 1 — build the payload (what you want to send)
    payload := map[string]any{
        "model": "gpt-4o",
        "messages": []map[string]string{
            {"role": "user", "content": "Hello!"},
        },
    }

    // Step 2 — serialize to JSON bytes
    body, _ := json.Marshal(payload)

    // Step 3 — construct the HTTP request
    req, _ := http.NewRequest("POST",
        "https://api.openai.com/v1/chat/completions",
        bytes.NewBuffer(body),
    )

    // Step 4 — set headers
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer sk-your-key")

    // Step 5 — execute
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    fmt.Println("Status:", resp.Status)
}
```

Key method: http.Client.Do(req) — this is the line that actually sends the request over the network. Everything before it is preparation.

2. Streaming
What it is
Instead of waiting for the entire response to be generated before sending it to you, the server sends tokens one by one as they are produced. This is how ChatGPT shows text appearing word by word.
Why it matters for AI
LLMs can take 10-30 seconds to generate a long response. Without streaming, your CLI freezes and the user stares at a blank screen. With streaming, they see output immediately.
How the protocol works
The API uses Server-Sent Events (SSE) — the HTTP connection stays open and the server sends:

How it works in Go
```
import (
    "bufio"
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
)

func streamChat(req *http.Request) error {
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // bufio.Scanner reads the response body LINE BY LINE
    // rather than waiting for the whole body to arrive
    scanner := bufio.NewScanner(resp.Body)

    for scanner.Scan() {
        line := scanner.Text()

        // SSE lines start with "data: "
        if !strings.HasPrefix(line, "data: ") {
            continue
        }

        // strip the "data: " prefix
        data := strings.TrimPrefix(line, "data: ")

        // [DONE] signals the stream is finished
        if data == "[DONE]" {
            break
        }

        // parse the JSON chunk
        var chunk struct {
            Choices []struct {
                Delta struct {
                    Content string `json:"content"`
                } `json:"delta"`
            } `json:"choices"`
        }

        if err := json.Unmarshal([]byte(data), &chunk); err != nil {
            continue
        }

        // print each token as it arrives — no newline, so it
        // appears to stream on the same line
        if len(chunk.Choices) > 0 {
            fmt.Print(chunk.Choices[0].Delta.Content)
        }
    }

    fmt.Println() // final newline
    return scanner.Err()
}
```

Key type: bufio.Scanner — wraps the response body and reads it incrementally line by line without loading everything into memory.

3. JSON Decoding
What it is
Converting raw JSON bytes (what the API returns) into Go structs (typed values your code can work with).
Two approaches

```
// Approach 1 — json.Unmarshal
// Use when you have the entire body already in memory
body, _ := io.ReadAll(resp.Body)

var result struct {
    Choices []struct {
        Message struct {
            Content string `json:"content"`
        } `json:"message"`
    } `json:"choices"`
}
json.Unmarshal(body, &result)
fmt.Println(result.Choices[0].Message.Content)


// Approach 2 — json.Decoder (preferred for HTTP responses)
// Use when reading directly from a stream/reader
var result ChatResponse
json.NewDecoder(resp.Body).Decode(&result)
// reads from resp.Body without loading it all into memory first
```
The json struct tags
```
type Message struct {
    Role    string `json:"role"`     // maps Go field "Role" to JSON key "role"
    Content string `json:"content"`
}
```
Key distinction: Unmarshal needs all bytes first. Decoder reads from a stream — better for large responses and streaming.

4. Context Cancellation
What it is
A mechanism to cancel an in-progress operation — closing the HTTP connection and freeing resources — when the user presses Ctrl+C, a timeout expires, or a parent operation is cancelled.
Why it matters for AI
LLM calls can take 30+ seconds. Without cancellation:

User presses Ctrl+C — your program exits but the HTTP request keeps running in the background, consuming goroutines and connections
A downstream service times out — your call hangs forever

How it works in Go

```
import (
    "context"
    "fmt"
    "net/http"
    "time"
)

func main() {
    // Option A — cancel after a fixed duration (timeout)
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel() // always call cancel to free resources

    // Option B — cancel when user presses Ctrl+C
    ctx, cancel = context.WithCancel(context.Background())
    // (you'd wire this to os.Signal handling)

    // Pass context INTO the request — Go will cancel the
    // HTTP connection automatically when ctx is done
    req, _ := http.NewRequestWithContext(ctx, "POST",
        "https://api.openai.com/v1/chat/completions", body)

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        // if ctx was cancelled, this error wraps context.Canceled
        fmt.Println("request failed or cancelled:", err)
        return
    }
    defer resp.Body.Close()
}
```
Key concept: ctx.Done() returns a channel that is closed when the context is cancelled. Goroutines check <-ctx.Done() to know when to stop.


5. Error Handling
What it is
AI APIs return structured error responses — not just Go errors. You need to handle both.
Two layers of errors

```
// Layer 1 — Go network error (DNS failure, timeout, connection refused)
resp, err := client.Do(req)
if err != nil {
    // network-level failure, no HTTP response
    return fmt.Errorf("network error: %w", err)
}
defer resp.Body.Close()

// Layer 2 — API error (wrong API key, rate limit, model not found)
// The HTTP connection SUCCEEDED but the API returned an error payload
if resp.StatusCode != http.StatusOK {
    var apiErr struct {
        Error struct {
            Message string `json:"message"`
            Type    string `json:"type"`
            Code    string `json:"code"`
        } `json:"error"`
    }
    json.NewDecoder(resp.Body).Decode(&apiErr)
    return fmt.Errorf("API error %d: %s (code: %s)",
        resp.StatusCode,
        apiErr.Error.Message,
        apiErr.Error.Code,
    )
}
```

6. Authentication
What it is
Proving to the API server that you are allowed to make requests. Almost all AI APIs use a Bearer token in the Authorization header.
How it works
```
// Standard Bearer token (OpenAI, Anthropic, most others)
req.Header.Set("Authorization", "Bearer "+apiKey)

// Anthropic uses a different header name
req.Header.Set("x-api-key", apiKey)
req.Header.Set("anthropic-version", "2023-06-01")

// Ollama (local) — usually no auth needed
// but you can optionally add one if you expose it with a proxy

import "os"

// Read from environment variable
apiKey := os.Getenv("OPENAI_API_KEY")
if apiKey == "" {
    log.Fatal("OPENAI_API_KEY is not set")
}
```


7. Embeddings
What it is
Converting text into a vector (a list of floating-point numbers) that represents the meaning of the text in a mathematical space. Similar meanings produce vectors that are numerically close to each other.
Why it matters
Embeddings are the engine behind:

Semantic search ("find docs similar to this query")
RAG (Retrieval-Augmented Generation — giving an LLM access to your own data)
Deduplication, clustering, recommendations

```
type EmbeddingRequest struct {
    Input string `json:"input"`
    Model string `json:"model"`
}

type EmbeddingResponse struct {
    Data []struct {
        Embedding []float64 `json:"embedding"` // the vector itself
        Index     int       `json:"index"`
    } `json:"data"`
}

// Create the request
payload := EmbeddingRequest{
    Input: "Kubernetes operators are controllers that manage CRDs",
    Model: "text-embedding-3-small",
}
body, _ := json.Marshal(payload)

req, _ := http.NewRequest("POST",
    "https://api.openai.com/v1/embeddings",
    bytes.NewBuffer(body))
req.Header.Set("Authorization", "Bearer "+apiKey)
req.Header.Set("Content-Type", "application/json")

// Execute
resp, _ := http.DefaultClient.Do(req)
var result EmbeddingResponse
json.NewDecoder(resp.Body).Decode(&result)

// result.Data[0].Embedding is now a slice of ~1536 float64 values
fmt.Println("Vector dimensions:", len(result.Data[0].Embedding))
```


8. Chat (Chat Completions)
What it is
The core API endpoint — you send a list of messages (a conversation history) and get back the next assistant message.
The message structure
```
System message  → sets the assistant's persona/behaviour
User message    → what the human said
Assistant message → what the AI replied (used to maintain history)
```

```
type ChatRequest struct {
    Model    string    `json:"model"`
    Messages []Message `json:"messages"`
    Stream   bool      `json:"stream,omitempty"`
}

type Message struct {
    Role    string `json:"role"`    // "system", "user", "assistant"
    Content string `json:"content"`
}

type ChatResponse struct {
    Choices []struct {
        Message Message `json:"message"`
    } `json:"choices"`
    Usage struct {
        PromptTokens     int `json:"prompt_tokens"`
        CompletionTokens int `json:"completion_tokens"`
        TotalTokens      int `json:"total_tokens"`
    } `json:"usage"`
}

// Multi-turn conversation — the history IS the memory
history := []Message{
    {Role: "system", Content: "You are a helpful Go programming assistant."},
    {Role: "user",   Content: "What is a goroutine?"},
    // after the first response, you append it:
    {Role: "assistant", Content: "A goroutine is a lightweight..."},
    // then the next user message:
    {Role: "user",   Content: "How is it different from a thread?"},
}

payload := ChatRequest{
    Model:    "gpt-4o",
    Messages: history,
}
```

### Detailed Explanation of AI **API** Integration

Detailed Explanation of AI **API** Integration Topics in GoThis guide covers the core topics required to build or maintain a production-grade AI orchestration engine in Go.1. **HTTP** Request ConstructionWhat it isThe foundation of everything. Every AI **API** (OpenAI, Anthropic, Ollama) is just an **HTTP** server. You send a request, you get a response. No magic.Why it matters for AIYou are not using a high-level **SDK** — you are building one. Every call to Claude or OpenAI is ultimately:**HTTPPOST** [https://api.openai.com/v1/chat/completions](https://api.openai.com/v1/chat/completions) Content-Type: application/json Authorization: Bearer sk-...

{ *model*: *gpt-4o*, *messages*: [...] } How it works in GoGopackage main

import (
    *bytes*
    *encoding/json*
    *fmt*
    *net/http*
)

func main() {
    // Step 1 — build the payload (what you want to send)
    payload := map[string]any{
    *model*: *gpt-4o*,
    *messages*: []map[string]string{
    {*role*: *user*, *content*: *Hello!*},
    },
    }

    // Step 2 — serialize to **JSON** bytes
    body, _ := json.Marshal(payload)

    // Step 3 — construct the **HTTP** request
    req, _ := http.NewRequest(*POST*, 
    *[https://api.openai.com/v1/chat/completions*,](https://api.openai.com/v1/chat/completions",) 
    bytes.NewBuffer(body),
    )

    // Step 4 — set headers
    req.Header.Set(*Content-Type*, *application/json*)
    req.Header.Set(*Authorization*, *Bearer sk-your-key*)

    // Step 5 — execute
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
    panic(err)
    }
    defer resp.Body.Close()

    fmt.Println(*Status:*, resp.Status)
}
Key method: http.Client.Do(req) — this is the line that actually sends the request over the network. Everything before it is preparation.2. StreamingWhat it isInstead of waiting for the entire response to be generated before sending it to you, the server sends tokens one by one as they are produced. This is how ChatGPT shows text appearing word by word.Why it matters for AILLMs can take 10 to 30 seconds to generate a long response. Without streaming, your **CLI** freezes and the user stares at a blank screen. With streaming, they see output immediately.How the protocol worksThe **API** uses Server-Sent Events (**SSE**) — the **HTTP** connection stays open and the server sends:Plaintextdata: {*choices*:[{*delta*:{*content*:*Hello*}}]}
data: {*choices*:[{*delta*:{*content*:* there*}}]}
data: {*choices*:[{*delta*:{*content*:*!*}}]}
data: [**DONE**]
How it works in GoGoimport (
    *bufio*
    *encoding/json*
    *fmt*
    *net/http*
    *strings*
)

func streamChat(req *http.Request) error {
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
    return err
    }
    defer resp.Body.Close()

    // bufio.Scanner reads the response body **LINE** BY **LINE**
    // rather than waiting for the whole body to arrive
    scanner := bufio.NewScanner(resp.Body)
    for scanner.Scan() {
    line := scanner.Text()

    // **SSE** lines start with *data: *
    if !strings.HasPrefix(line, *data: *) {
    continue
    }

    // strip the *data: * prefix
    data := strings.TrimPrefix(line, *data: *)

    // [**DONE**] signals the stream is finished
    if data == *[**DONE**]* {
    break
    }

    // parse the **JSON** chunk
    var chunk struct {
    Choices []struct {
    Delta struct {
    Content string `json:*content*`
    } `json:*delta*`
    } `json:*choices*`
    }

    if err := json.Unmarshal([]byte(data), &chunk); err != nil {
    continue
    }

    // print each token as it arrives — no newline, so it
    // appears to stream on the same line
    if len(chunk.Choices) > 0 {
    fmt.Print(chunk.Choices[0].Delta.Content)
    }
    }

    fmt.Println() // final newline
    return scanner.Err()
}
Key type: bufio.Scanner — wraps the response body and reads it incrementally line by line without loading everything into memory.3. **JSON** DecodingWhat it isConverting raw **JSON** bytes (what the **API** returns) into Go structs (typed values your code can work with).Two approachesGo// Approach 1 — json.Unmarshal
// Use when you have the entire body already in memory
body, _ := io.ReadAll(resp.Body)
var result struct {
    Choices []struct {
    Message struct {
    Content string `json:*content*`
    } `json:*message*`
    } `json:*choices*`
}
json.Unmarshal(body, &result)
fmt.Println(result.Choices[0].Message.Content)

// Approach 2 — json.Decoder (preferred for **HTTP** responses)
// Use when reading directly from a stream/reader
var result ChatResponse
json.NewDecoder(resp.Body).Decode(&result) 
// reads from resp.Body without loading it all into memory first
The json struct tagsGotype Message struct {
    Role    string `json:*role*`     // maps Go field *Role* to **JSON** key *role*
    Content string `json:*content*`
}
Key distinction: Unmarshal needs all bytes first. Decoder reads from a stream — better for large responses and streaming.4. Context CancellationWhat it isA mechanism to cancel an in-progress operation — closing the **HTTP** connection and freeing resources — when the user presses Ctrl+C, a timeout expires, or a parent operation is cancelled.Why it matters for **AILLM** calls can take 30+ seconds. Without cancellation:User presses Ctrl+C — your program exits but the **HTTP** request keeps running in the background, consuming goroutines and connections.A downstream service times out — your call hangs forever.How it works in GoGoimport (
    *context*
    *fmt*
    *net/http*
    *time*
)

func main() {
    // Option A — cancel after a fixed duration (timeout)
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel() // always call cancel to free resources

    // Option B — cancel when user presses Ctrl+C
    // ctx, cancel = context.WithCancel(context.Background())
    // (you'd wire this to os.Signal handling)

    // Pass context **INTO** the request — Go will cancel the
    // **HTTP** connection automatically when ctx is done
    req, _ := http.NewRequestWithContext(ctx, *POST*, 
    *[https://api.openai.com/v1/chat/completions*,](https://api.openai.com/v1/chat/completions",) body)

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
    // if ctx was cancelled, this error wraps context.Canceled
    fmt.Println(*request failed or cancelled:*, err)
    return
    }
    defer resp.Body.Close()
}
Key concept: ctx.Done() returns a channel that is closed when the context is cancelled. Goroutines check <-ctx.Done() to know when to stop.5. Error HandlingWhat it isAI APIs return structured error responses — not just Go errors. You need to handle both.Two layers of errorsGo// Layer 1 — Go network error (**DNS** failure, timeout, connection refused)
resp, err := client.Do(req)
if err != nil {
    // network-level failure, no **HTTP** response
    return fmt.Errorf(*network error: %w*, err)
}
defer resp.Body.Close()

// Layer 2 — **API** error (wrong **API** key, rate limit, model not found)
// The **HTTP** connection **SUCCEEDED** but the **API** returned an error payload
if resp.StatusCode != http.StatusOK {
    var apiErr struct {
    Error struct {
    Message string `json:*message*`
    Type    string `json:*type*`
    Code    string `json:*code*`
    } `json:*error*`
    }
    json.NewDecoder(resp.Body).Decode(&apiErr)
    return fmt.Errorf(***API** error %d: %s (code: %s)*, 
    resp.StatusCode, 
    apiErr.Error.Message, 
    apiErr.Error.Code,
    )
}
Common **HTTP** status codes from AI APIsCodeMeaningAction200SuccessContinue400Bad request (wrong params)Fix your payload401Invalid **API** keyCheck key429Rate limit exceededRetry with backoff500Server errorRetry503Service unavailableRetry6. AuthenticationWhat it isProving to the **API** server that you are allowed to make requests. Almost all AI APIs use a Bearer token in the Authorization header.How it worksGo// Standard Bearer token (OpenAI, Anthropic, most others)
req.Header.Set(*Authorization*, *Bearer *+apiKey)

// Anthropic uses a different header name req.Header.Set(*x-api-key*, apiKey) req.Header.Set(*anthropic-version*, ***2023**-06-01*)

// Ollama (local) — usually no auth needed // but you can optionally add one if you expose it with a proxy Best practice — never hardcode keysGoimport *os*

// Read from environment variable
apiKey := os.Getenv(*OPENAI_API_KEY*)
if apiKey == "* {
    log.Fatal(*OPENAI_API_KEY is not set")
}
## Embeddings What it isConverting text into a vector (a list of floating-point numbers) that represents the meaning of the text in a mathematical space. Similar meanings produce vectors that are numerically close to each other.Why it mattersEmbeddings are the engine behind:Semantic search (*find docs similar to this query*)RAG (Retrieval-Augmented Generation — giving an LLM access to your own data)Deduplication, clustering, recommendationsHow a request looksGotype EmbeddingRequest struct {
    Input string `json:*input*`
    Model string `json:*model*`
}

type EmbeddingResponse struct {
    Data []struct {
    Embedding []float64 `json:*embedding*` // the vector itself
    Index     int       `json:*index*`
    } `json:*data*`
}

// Create the request
payload := EmbeddingRequest{
    Input: "Kubernetes operators are controllers that manage CRDs*,
    Model: *text-embedding-3-small",
}
body, _ := json.Marshal(payload)

req, _ := http.NewRequest(*POST*, 
    *[https://api.openai.com/v1/embeddings*,](https://api.openai.com/v1/embeddings*,) 
    bytes.NewBuffer(body))

req.Header.Set(*Authorization*, *Bearer *+apiKey) req.Header.Set(*Content-Type*, *application/json")

// Execute resp, _ := http.DefaultClient.Do(req) var result EmbeddingResponse json.NewDecoder(resp.Body).Decode(&result)

// result.Data[0].Embedding is now a slice of ~**1536** float64 values
fmt.Println(*Vector dimensions:*, len(result.Data[0].Embedding))
## Chat (Chat Completions)What it isThe core API endpoint — you send a list of messages (a conversation history) and get back the next assistant message.The message structureSystem message: sets the assistant's persona/behaviourUser message: what the human saidAssistant message: what the AI replied (used to maintain history)How a request looks in GoGotype ChatRequest struct {
    Model    string    `json:*model*`
    Messages []Message `json:*messages*`
    Stream   bool      `json:*stream,omitempty*`
}

type Message struct {
    Role    string `json:*role*` // *system*, *user*, *assistant*
    Content string `json:*content*`
}

type ChatResponse struct {
    Choices []struct {
    Message Message `json:*message*`
    } `json:*choices*`
    Usage struct {
    PromptTokens     int `json:*prompt_tokens*`
    CompletionTokens int `json:*completion_tokens*`
    TotalTokens      int `json:*total_tokens*`
    } `json:*usage*`
}

// Multi-turn conversation — the history IS the memory
history := []Message{
    {Role: *system*, Content: *You are a helpful Go programming assistant.*},
    {Role: *user*,   Content: *What is a goroutine?*},
    // after the first response, you append it:
    {Role: *assistant*, Content: *A goroutine is a lightweight...*},
    // then the next user message:
    {Role: *user*,   Content: *How is it different from a thread?*},
}

payload := ChatRequest{
    Model:    *gpt-4o*,
    Messages: history,
}
## Tool Calling (Function Calling)What it isYou define functions/tools that the LLM can *call* — it decides when to use them and returns structured arguments you then execute on its behalf. This is how AI agents work.The flowYou send: messages + list of available toolsLLM responds with: "call tool get_weather with args {city: 'Bangalore'}"Your code: actually calls the get_weather functionYou send back: the result of that callLLM: uses the result to write a final responseHow it looks in GoGo// Step 1 — define the tools in your request
type Tool struct {
    Type     string       `json:*type*`
    Function ToolFunction `json:*function*`
}

type ToolFunction struct {
    Name        string         `json:*name*`
    Description string         `json:*description*`
    Parameters  map[string]any `json:*parameters*` // **JSON** Schema
}

tools := []Tool{{
    Type: *function*,
    Function: ToolFunction{
    Name:        *get_weather*,
    Description: *Get current weather for a city*,
    Parameters: map[string]any{
    *type*: *object*,
    *properties*: map[string]any{
    *city*: map[string]any{
    *type*:        *string*,
    *description*: *The city name*,
    },
    },
    *required*: []string{*city*},
    },
    },
}}

// Step 2 — check if the **LLM** wants to call a tool var response ChatResponse json.NewDecoder(resp.Body).Decode(&response)

choice := response.Choices[0]
if choice.FinishReason == *tool_calls* {
    // **LLM** decided to call a tool — extract the arguments
    toolCall := choice.Message.ToolCalls[0]
    var args map[string]string
    json.Unmarshal([]byte(toolCall.Function.Arguments), &args)

    // Step 3 — execute the actual function in your code
    result := getWeather(args[*city*])

    // Step 4 — send the result back so the **LLM** can use it
    history = append(history, choice.Message) // the tool_call message
    history = append(history, Message{
    Role:       *tool*,
    Content:    result,
    ToolCallID: toolCall.ID,
    })
    // make another **API** call with the updated history...
}
## File UploadsWhat it isSome AI APIs accept files directly — PDFs, images, audio. The standard way to send binary files over HTTP is multipart form data, the same mechanism HTML forms use.How it works in GoGoimport (
    *bytes*
    *fmt*
    *io*
    *mime/multipart*
    *net/http*
    *os*
    *path/filepath*
)

func uploadFile(apiKey, filePath string) error {
    // Open the file
    file, err := os.Open(filePath)
    if err != nil {
    return err
    }
    defer file.Close()

    // Create a multipart writer — it builds the multipart body
    var body bytes.Buffer
    writer := multipart.NewWriter(&body)

    // Add the *purpose* text field (required by OpenAI Files **API**)
    writer.WriteField(*purpose*, *assistants*)

    // Add the file field
    part, err := writer.CreateFormFile(*file*, filepath.Base(filePath))
    if err != nil {
    return err
    }
    io.Copy(part, file) // copy file contents into the multipart body

    // **IMPORTANT**: close the writer to write the final boundary
    writer.Close()

    // Build the request
    req, _ := http.NewRequest(*POST*, 
    *[https://api.openai.com/v1/files*,](https://api.openai.com/v1/files*,) 
    &body,
    )
    req.Header.Set(*Authorization*, *Bearer "+apiKey)
    
    // multipart content type includes the boundary marker
    req.Header.Set(*Content-Type*, writer.FormDataContentType())

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
    return err
    }
    defer resp.Body.Close()

    fmt.Println(*Upload status:*, resp.Status)
    return nil
}
Key type: multipart.Writer — manages the boundary markers that separate different parts (fields and files) in the request body.How They All Connect TogetherEvery AI feature you build uses all of these pieces in unison:PlaintextFile uploads ──────────────────────────────────────┐
```
File uploads ──────────────────────────────────────┐
Authentication ─────────────────────────────────┐  │
Context cancellation ────────────────────────┐  │  │
Error handling ───────────────────────────┐  │  │  │
                                          │  │  │  │
HTTP request construction                 │  │  │  │
       ↓                                  │  │  │  │
[headers: Auth ───────────────────────────┘  │  │  │
 body: JSON ← JSON encoding                  │  │  │
 ctx: cancel ← Context ──────────────────────┘  │  │
 multipart ← File upload ───────────────────────┘  │
]                                                  │
       ↓                                           │
client.Do(req) ────────────────────────────────────┘
       ↓
Error check (status code + API error body)
       ↓
┌──────────────────────────────┐
│ Stream? → bufio.Scanner      │  ← Streaming
│ No?     → json.Decoder       │  ← JSON decoding
└──────────────────────────────┘
       ↓
Structured result 
(Chat response / Embedding vector / Tool call arguments)
```
Every single topic is a critical step in this production pipeline. Once you orchestrate these components together, you can interface seamlessly with any modern AI platform.