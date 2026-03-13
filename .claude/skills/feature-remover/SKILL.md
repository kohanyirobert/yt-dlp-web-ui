# feature-remover

Removes a feature from the codebase based on a pre-written plan.

## Instructions

You are a feature remover. Follow these steps exactly:

### Step 1: Require a plan file

Check the `.claude/plans/` directory for plan files. If the user specified a plan file path, use that. Otherwise, list available plans and ask the user which one to use.

If there are NO plan files in `.claude/plans/` and the user did not specify one, **stop immediately** with this message:
> No plan found. Please create a plan first (e.g. using plan mode) before running /feature-remover.

### Step 2: Read the plan

Read the selected plan file completely. Understand what feature is being removed and what changes are required.

### Step 3: Create a branch

Create a new git branch from the current branch using the naming convention:
```
feature/remove-<feature-name>
```
where `<feature-name>` is a kebab-case summary of the feature being removed, derived from the plan.

### Step 4: Implement the removal

Implement all changes described in the plan. After making changes:
- Run `go test ./...` to verify backend changes compile and pass tests
- Run `cd frontend && pnpm build` to verify frontend changes compile (if frontend was modified)
- Fix any issues found during verification

### Step 5: Ship it — commit, merge, and land on main

Stage all changed files and create a single commit with a descriptive message summarizing the removal. Use the format:
```
Remove <feature name> (#<PR number will be added by GitHub>)

<brief description of what was removed and why>
```

Then execute the full landing sequence:

1. **Push** the branch to the remote: `git push --set-upstream origin <branch-name>`
2. **Open a PR** using `gh pr create` targeting the default branch with a clear title and body
3. **Wait for CI** — poll the PR checks with `gh pr checks <PR number> --watch` until they pass. If checks fail, diagnose, fix, commit, and push again. Repeat until green.
4. **Merge the PR** — once checks pass, squash-merge with `gh pr merge <PR number> --squash --delete-branch`
5. **Come home** — switch back to `main` with `git checkout main` and pull the merged changes with `git pull` so the local branch is fully up to date with the remote

After landing on main, **stop completely**. Do NOT continue with any other work.
