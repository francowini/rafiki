# Bruno Basics

## What is Bruno?

Bruno is like Postman - an API testing tool. The main difference is that Bruno stores your API requests as plain text files in your Git repository.

**Key Benefits:**
- Requests stored as `.bru` files in your project
- Works offline (no cloud sync)
- Free and open source
- Easy to version control with Git

## Installation

```bash
# macOS
brew install bruno

# Windows
winget install Bruno.Bruno

# Or download from: https://www.usebruno.com/downloads
```

## How to Use

### 1. Open Collection
1. Open Bruno app
2. File → Open Collection
3. Navigate to `/Users/francowini/Documents/rafiki/bruno/rafiki`
4. Click "Open"

### 2. Select Environment
- Click dropdown at top (shows "No Environment")
- Select **"local"** for local testing
- This sets `base-url` to `http://localhost:3000`

### 3. Send a Request
1. Click any request in the left sidebar
2. Click **"Send"** button
3. View response below

## Request Structure

Bruno requests are stored as `.bru` files with this format:

```
meta {
  name: Create Think
  type: http
}

post {
  url: {{base-url}}/v1/thinks
  body: json
  auth: bearer
}

headers {
  Content-Type: application/json
}

body:json {
  {
    "category": "reflection",
    "content": "My think"
  }
}

auth:bearer {
  token: {{jwt-token}}
}
```

## Environment Variables

Variables are defined in `environments/local.bru`:

```
vars {
  base-url: http://localhost:3000
  jwt-token: your-token-here
}
```

Use variables in requests with `{{variable-name}}`.

## Setting Variables from Response

You can extract values from responses and save them:

```javascript
tests {
  const data = res.getBody();
  bru.setEnvVar("jwt-token", data.token);
}
```

## Authentication Types

### No Auth
```
get {
  auth: none
}
```

### Basic Auth (for login)
```
get {
  auth: basic
}

auth:basic {
  username: test@example.com
  password: password123
}
```

### Bearer Token (for API requests)
```
get {
  auth: bearer
}

auth:bearer {
  token: {{jwt-token}}
}
```

That's it! Bruno is simple and lightweight - just open the collection and start testing.
