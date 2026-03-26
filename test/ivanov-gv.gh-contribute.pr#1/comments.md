issue #4038597073 by @ai-contributor-helper | hidden: Resolved

---
issue #4038819817 by you (@ivanov-gv) | hidden: Resolved

---
review #3929204495 by you (@ivanov-gv) | hidden: Resolved

---
review #3929240428 by you (@ivanov-gv)  
_2026-03-11 12:24:52_  

comments: 1  

---
review #3929353771 by you (@ivanov-gv) | hidden: Resolved

---
issue #4039142865 by @ai-contributor-helper | hidden: Resolved

---
issue #4039221478 by @ai-contributor-helper  
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

---
review #3929758963 by you (@ivanov-gv)  
_2026-03-11 13:51:01_  

>7. A review by @ivanov-gv 
>
>a few fixes are needed as well as some further feature implementations. looks good so far! 

comments: 1  

---
issue #4039593663 by @ai-contributor-helper  
_2026-03-11 14:25:31_  

>8. A markdown comment from the bot
>
>## Review feedback addressed
>
>All items from review #3929758963 resolved:
>
>- **Formatting**: removed `#` prefix, added `---` separators, trailing spaces for md line breaks, renamed "by you" → "reactions by you"
>- **Timeline**: issue comments and reviews merged and sorted by date
>- **PR format**: `===` → `---`
>- **CLI positional args**: `pr 1` and `comments <id>` work without flags
>- **New command**: `review <id>` — shows inline comments with thread grouping (replies indented with `>`)

---
review #3930039277 by you (@ivanov-gv)  
_2026-03-11 14:31:04_  

>9. A review from @ivanov-gv 
>more changes are needed

comments: 3  

---
issue #4041153603 by @ai-contributor-helper  
_2026-03-11 18:11:44_  

>10. A comment from the bot
>
>## Review #3930039277 addressed
>
>- **Conversation count**: `pr` output now shows `Conversation: N comments` after Issues
>- **Review hidden detection**: hide if all comments minimized OR all linked threads resolved — checks all thread replies too, so review #3929240428 (reply-only) is now correctly hidden
>- **`review` command**: shows comment IDs, file+line location, reactions; `---` separator before comments; no diff hunk (commit hash not accessible via GitHub App)

---
issue #4042410800 by @ai-contributor-helper  
_2026-03-11 21:42:18_  

>11. A comment from the bot with eyes emoji
>refactoring test

(1 👀)  
reactions by you:  

---
issue #4067633036 by you (@ivanov-gv)  
_2026-03-16 13:22:54_  

>12. A comment by @ivanov-gv posted with the bot.
>test comment

