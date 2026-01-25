# Quick Commands

docker system prune

docker builder prune

docker image prune -a

## Prompts

### PR Prompt
How should I PR my changes? Give me a logical progression and the exact commands including git checkout git pull etc. Give great commit and PR messages but don't mention that claude was involved. Don't stash anything or delete any files.

### Commit Prompt
How should I commit my changes? Give me a logical progression and the exact commands including git checkout git pull etc. Give great commit messages but don't mention that claude was involved. Don't stash anything or delete any files.

### App Review Prompt
Take a look at my @src/app/  what are 3 things you would fix, what 3 things you would refactor, what 3 things you would take away, what 3 features should I add, what 3 things can I do to more deeply integrate my features? What 5 things out of all those should I do now for the highest return?

### Frontend Review Prompt
Can you make the UI UX and QoL better on my @frontend/app/ stuff? Make sure you add or update tests.

Take a look at my project what are 3 things you would fix, what 3 things you would refactor, what 3 things you would take away, what 3 features should I add, what 3 things can I do to more deeply integrate my features? What 5 things out of all those should I do now for the highest return? Double check that none of the things you are suggestion are already done.

---

## Database Volume Best Practices (IMPLEMENTED ✅)

### What Was Fixed

1. ✅ Separate volumes: `postgres_data` (dev) vs `postgres_data_prod` (prod)
2. ✅ Backup service now correctly mounts `postgres_data_prod` (was `postgres_data`)
3. ✅ Prod compose uses `${VAR:?error}` syntax - fails fast if credentials not set
4. ✅ Prod defaults to `DB_SSLMODE=require` instead of `disable`
5. ✅ Test uses tmpfs (no persistent volume needed)

### Volume Strategy

| Environment | Volume Name          | Credentials                         |
| ----------- | -------------------- | ----------------------------------- |
| Dev         | `postgres_data`      | devuser/devpass/starter_kit_db      |
| Test        | tmpfs (ephemeral)    | testuser/testpass/starter_kit_test  |
| Prod        | `postgres_data_prod` | **REQUIRED** - no defaults          |

### Key Points

- Postgres auto-initializes DB/user from `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD` on first boot
- If volume has data, these env vars are ignored
- App handles schema via migrations (not manual SQL)
- Never share volumes between environments
