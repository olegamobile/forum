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
erDiagram
    USER {
        user_id TEXT "*PK, UUID"
        email TEXT
        username TEXT
        password TEXT
        created_at TEXT
    }
    POST {
        post_id INT "*PK"
        user_id TEXT "UUID"
        title TEXT
        content TEXT
        created_at TEXT
        categories TEXT "JSON"
        parent_id INT
        base_id INT
    }
    REACTION {
        id INT "*PK"
        user_id TEXT "UUID"
        post_id INT
        reaction_type TEXT
        created_at TEXT
    }
    SESSION {
        id INT "*PK"
        user_id TEXT "UUID"
        session_token TEXT
        expires_at TEXT
    }

    USER ||--|| SESSION : has
    USER ||--o{ POST : creates
    USER ||--o{ REACTION : gives
    POST ||--o{ REACTION : receives

```

## Members
<...>