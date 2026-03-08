# Fork Workflow Guide

## Repository Structure

```
origin    → https://github.com/brendantza/picoclaw.git (your fork)
upstream  → https://github.com/sipeed/picoclaw.git (original repo)
```

## Branch Strategy

| Branch | Purpose | Sync With |
|--------|---------|-----------|
| `main` | Mirror of upstream | `upstream/main` |
| `development` | Your working branch | `origin/development` |
| `feature/*` | Feature branches (optional) | - |

## Daily Workflow

### 1. Start Work
```bash
git checkout development
# make changes
git add .
git commit -m "your changes"
git push origin development
```

### 2. Sync with Upstream
```bash
# Easy way (using the script)
git sync

# Or manual merge way
git fetch upstream
git checkout main
git merge upstream/main
git push origin main
git checkout development
git merge main
git push origin development
```

### 3. Handle Conflicts

If conflicts occur during sync:

```bash
# 1. See which files conflict
git status

# 2. Edit conflicted files, look for 
#    <<<<<<< HEAD
#    your changes
#    =======
#    upstream changes
#    >>>>>>> main

# 3. After editing, mark resolved
git add <resolved-files>
git rebase --continue   # if rebasing
git merge --continue    # if merging

# 4. Push
git push origin development
```

## Best Practices

1. **Never commit directly to `main`** - Keep it as a clean mirror of upstream
2. **Always work on `development` or feature branches**
3. **Sync regularly** - Small, frequent syncs are easier than large ones
4. **Test after syncing** - Ensure your changes still work

## Pull Request Workflow

To contribute back to upstream:

```bash
# 1. Sync your main
git checkout main
git fetch upstream
git reset --hard upstream/main
git push origin main --force

# 2. Create feature branch from clean main
git checkout -b feature/my-contribution

# 3. Make changes and push
git add .
git commit -m "feat: description"
git push origin feature/my-contribution

# 4. Create PR via GitHub UI from brendantza/picoclaw to sipeed/picoclaw
```
