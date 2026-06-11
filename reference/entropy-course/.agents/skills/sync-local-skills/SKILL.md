---
name: sync-local-skills
description: Detect repo-local skills under .agents/skills, inventory their SKILL.md metadata, and update AGENTS.md with a ## Local Skills section containing `${skill-name}` tokens and concise situational triggers. Use when asked to list, sync, append, refresh, or maintain local skills in project agent instructions.
---

# Sync Local Skills

Use this skill to refresh project agent instructions so they mention every repo-local skill.

## Workflow

1. Run the inventory helper from the repository root:

   ```bash
   node .agents/skills/sync-local-skills/scripts/inventory-local-skills.mjs --markdown
   ```

2. Update `AGENTS.md`:
   - Replace an existing `## Local Skills` section if present.
   - Otherwise insert `## Local Skills` near other agent behavior rules, preferably after architecture/design guidance and before test-running guidance.
   - Preserve unrelated `AGENTS.md` sections and wording.

3. Write the section with this shape:

   ```markdown
   ## Local Skills

   When a task matches a local skill, load that skill's `SKILL.md` before acting and follow only the relevant referenced material. Use `.agents/skills/<name>` directory names as the canonical skill tokens.

   - `${skill-name}`: Use for concise situational trigger text.
   ```

4. Validate coverage:

   ```bash
   node .agents/skills/sync-local-skills/scripts/inventory-local-skills.mjs --check-agents
   ```

## Trigger Wording

- Use `.agents/skills/<name>` directory names as canonical tokens, not frontmatter `name` fields.
- Include every `.agents/skills/*/SKILL.md`, including this skill when it exists in the repo.
- Derive each bullet from the skill frontmatter `description`, but tighten long descriptions into one concise trigger.
- Prefer `Use for ...` or `Use when ...` phrasing.
- Do not paste very long descriptions into `AGENTS.md` when a shorter situational trigger preserves the intent.

## Helper Script

`scripts/inventory-local-skills.mjs` is read-only. It inventories local skills, prints their frontmatter descriptions, and checks whether `AGENTS.md` contains every `${skill-name}` token.
