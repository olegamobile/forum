# Forum

A simple forum using SQLite and Go.
<...>

<!-- ### Installing gcc on Windows


Here's a good tutorial:
https://code.visualstudio.com/docs/cpp/config-mingw
 -->

## Installation

1. Make sure the following are installed in your sytem before running the program:
- Go (version 1.20 or later recommended)
- Git, such as Gitea, to clone the repository

2. Clone the repository by running the following code:
<...>

## Usage

You can now use the application by running the following command:
<...>

## Implementation

Following is an entity relationship diagram (ERD) showing the relationships among entities with their corresponding attributes:

```mermaid
---
config:
  theme: forest
title: FORUM
---

erDiagram
   USER {
        *user_id TEXT(UUID)
        email TEXT
        username TEXT
        password TEXT
        created_at TEXT
    }
    P["POST/COMMENT"] {
        *post_id INT
        user_id TEXT(UUID)
        title TEXT
        content TEXT
        created_at TEXT
        categories TEXT(JSON)
        parent_id INT
        base_id INT
    }
    R[REACTION] {
        *id INT
        user_id TEXT(UUID)
        post_id INT
        reaction_type TEXT
        created_at TEXT
    }
    S[SESSION] {
        *id INT
        user_id TEXT(UUID)
        session_token TEXT
        expires_at TEXT
    }

    USER ||--|| S : has
    USER ||--o{ P : has
    USER ||--o{ R : has
    P ||--o{ R : has
```

## Members
<...>