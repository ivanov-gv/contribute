---
name: review-cycle
description: Respond to PR review feedback — read reviews, address comments, reply, resolve threads
user_invocable: true
---

# Review Cycle Skill

You are an AI agent responding to PR review feedback using `gh-contribute`.

## Usage

```
/review-cycle [PR number]
```

If no PR number is given, auto-detect from the current branch.

## Workflow

### 1. Read the current PR state

```bash
gh-contribute comments --pr <N>
```

Look for new reviews (state: CHANGES_REQUESTED or COMMENTED) that have not been addressed yet.

### 2. For each unaddressed review

```bash
gh-contribute review <review-id> --pr <N>
```

Read all inline threads in the review.

### 3. For each thread in the review

1. **Acknowledge** — react with 👀 to signal you've seen it:
   ```bash
   gh-contribute react <comment-id> EYES
   ```

2. **Understand** — read the comment body and the file/line context.

3. **Fix** — make the requested code change. Commit with a message referencing the thread.

4. **Reply** — tell the reviewer what you did:
   ```bash
   gh-contribute reply <comment-id> "Fixed in <short-sha> — <what you changed>"
   ```

5. **React** — signal completion:
   ```bash
   gh-contribute react <comment-id> ROCKET
   ```

6. **Resolve** — mark the thread as resolved:
   ```bash
   gh-contribute resolve <thread-id> --pr <N>
   ```

### 4. Push and notify

```bash
git push
gh-contribute comment "All feedback addressed, PTAL" --pr <N>
```

### 5. Check for approval

```bash
gh-contribute comments --pr <N>
```

If the latest review is APPROVED, the cycle is complete. Otherwise, wait for the next review.

## Important Notes

- Always read the full review before making changes — understand the overall feedback first.
- Group related fixes into a single commit when possible.
- If a comment is unclear, reply asking for clarification instead of guessing.
- Never force-push during a review cycle — it breaks the review context.
- Use `gh-contribute thread <thread-id>` to see the full conversation history of a specific thread.
