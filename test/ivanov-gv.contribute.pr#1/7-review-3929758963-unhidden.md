review #3929758963 by @ivanov-gv  
_2026-03-11 13:51:01_

>7. A review by @ivanov-gv 
>
>a few fixes are needed as well as some further feature implementations. looks good so far! 
>
>
>thread #2918508377  TODO.md on line +48  
>comment #2918508377 by @ivanov-gv  
>_2026-03-11 13:51:01_
>
>7.1 A conversation with a huge markdown body. Has a rocket and eyes reactions by the bot. TODO.md on line R48
>
>let's change the pattern a bit.
>
>comments:
>- no # before issue, instead - a separator line `---`
>- don't forget to add `  ` (two spaces) at the end of the `issue ...` lines, so md renders draw date on a following line
>- by you -> reactions by you - to make it obvious
>
>```
>issue #4038597073 by you (@ivanov-gv-ai-helper) | hidden: RESOLVED
>
>---
>issue #4038819817 by @ivanov-gv | hidden: RESOLVED
>
>---
>issue #4039142865 by you (@ivanov-gv-ai-helper) | hidden: RESOLVED  
>
>---  
>issue #4039221478 by you (@ivanov-gv-ai-helper)  
>_2026-03-11 13:29:09_
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
>
>---
>```
>
>additional comment on reviews:
>- all the reviews below are actually hidden. check their status
>- reviews must be sorted with issues by date, forming a timeline
>
>```
>review #3929204495 by @ivanov-gv  
>_2026-03-11 12:17:34_
>
>submit review
>
>comments: 2  
>(1 👀)  
>reactions by you:  
>
>---
>review #3929240428 by @ivanov-gv  
>_2026-03-11 12:24:52_
>
>comments: 1
>
>---
>review #3929353771 by @ivanov-gv  
>_2026-03-11 12:45:28_
>
>resolved review
>
>comments: 1  
>
>```
>
>pr:
>
>- instead of === use ---
>
>cli usage:
>- make `gh-contribute pr` take a pr id as an argument without --pr key. so `gh-contribute pr --pr 1` becomes `gh-contribute pr 1`
>- add comment id to `comments` the same way for showing only one comment: `gh-contribute comments 4039221478`
>
>after resolving these, add a command for viewing a review. comments in a review should be sorted by date primarily. comment branches (a comment and its replies) should be printed together, a reply after its comment. I can't imagine how yet, but branches also should be viewed differently, to make the reader understand it's a chain of comments. 
>
>that's it. good luck!

(1 👀 1 🚀)  
reactions by you: (1 👀 1 🚀)  
