---
session_id: sess-aa6be2e687a248d5bcb0a7d26daac667
created_at: 2025-03-04T10:00:00Z
title: "Simple chat with code"
---

## System Instructions

You are a helpful programming assistant.
Be concise and provide clear examples.

## History

### 1. **User**
How do I create a simple HTTP server in Go?

### 2. **Assistant**
Here's a basic HTTP server in Go:

```go
package main

import (
	"fmt"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!")
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
```

Run this and visit `http://localhost:8080` to see it in action.

### 3. **User**
Thanks!

