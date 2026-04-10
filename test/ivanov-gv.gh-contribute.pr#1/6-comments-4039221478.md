issue #4039221478 by you (@ai-contributor-helper)  
_2026-03-11 13:29:09_  

>6. An unresolved comment by the bot with markdown inside
>
>## gh-contribute v1 — done
>
>Implemented the CLI extension with 4 commands:
>- `pr` — PR details in markdown (title, state, reviewers, assignees, labels, etc.)
>- `comments` — list issue comments and reviews with reactions, "by you" tracking, hidden/resolved detection
>- `comment` — post top-level PR comments
>- `react` — add reactions to comments
>
>Architecture:
>- GraphQL API v4 for read operations (rich nested data in fewer calls)
>- REST API v3 for write operations (post comment, add reaction)
>- Auto-detects repo from git remote, PR from current branch
>- Output is human-readable markdown, not JSON
