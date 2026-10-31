---
name: skill-creator
description: >-
  Create a new Agent Skill (a folder with a SKILL.md) that follows current best
  practice: correct frontmatter, a description that makes the skill easy to
  find, a concise body, and optional reference, script and asset files. Use when
  the user asks to create, write, or scaffold a skill, or to turn a repeated
  workflow into one.
---

# Skill Creator

## Goal

Produce one skill folder that an agent can find from its description, load
cheaply, and follow without guessing. The skill must work on its first real
task and be easy to improve afterwards.

## Steps

1. **Pick the task.** Choose a workflow the user has already done by hand at

 least once. Write down the context they had to supply each time: names, rules,
 paths, preferences. The skill should hold that context.

2. **Name it.** Use lowercase letters, digits and hyphens only, at most 64

 characters, with no leading, trailing or doubled hyphens. The folder name must
 equal `name`. Prefer a gerund phrase such as `processing-pdfs` or
 `reviewing-contracts`. Avoid vague names like `helper` or `utils`, and reserved
 words like `anthropic` or `claude`.

3. **Write the description** (1 to 1024 characters) in the third person. Say what

 the skill does and when to use it, using the words a user would actually say.
 Example: "Extracts text and tables from PDF files, fills forms and merges
 documents. Use when the user mentions PDFs, forms or document extraction."
 Avoid vague lines like "Helps with documents" or "Processes data".

4. **Write the body.** Include only what the agent does not already know:
  - the steps, in order, with the exact commands or paths to use;
  - one or two input/output examples;
  - the edge cases the user has already hit.
    Use forward slashes in every path. Give one default approach, not a list of
    alternatives. Match the freedom to the risk: exact commands for fragile steps
    such as migrations or deletions, guidance for judgment calls such as reviews or
    writing.
5. **Split when long.** Keep `SKILL.md` under 500 lines. Move detail into

 `references/` (docs, schemas, domain notes), `scripts/` (code the agent runs
 rather than rewrites) and `assets/` (templates). Link each file from `SKILL.md`
 one level deep. Put a table of contents at the top of any reference file over
 100 lines. Put information in one place only: in `SKILL.md` or in a reference
 file, never both.

6. **Add no extra documents.** Do not create a README, CHANGELOG, installation

 guide, quick reference, or notes about how the skill was made. The skill folder
 holds only what the agent needs for the task.

7. **Add the UI metadata.** Create `agents/openai.yaml` with `display_name`, a

 `short_description` of 25 to 64 characters, and a `default_prompt` that names
 the skill as `$<skill-name>`. Quote every string value. Add icons or
 dependencies only when the user supplies them.

8. **Let scripts handle their own errors.** Catch the failure cases in the script,

 comment each constant with its reason, and list the packages it needs.

9. **Check it.** If the reference library is installed, validate the frontmatter

 with `skills-ref validate <folder>`. Otherwise check the rules in steps 2 and 3
 by hand. Then run the skill on one real task that did not inform its writing,
 and compare the result with the same task done without the skill. Fix what the
 agent missed and run again. Use three test cases at minimum.

10. **Iterate from what you observe.** When the agent skips a step or reads the

wrong file, make the instruction more specific or move it higher in the file.
Do not add general advice the agent did not need.

## Protocols

### Clarify

- Always research best practice on the internet before answering your questions

or clarifying anything you are unsure about.

- Use `AskUserQuestion` if it is available. Ask as if the user is new to this

topic: plain words, one idea per question.

- Ask about the task's trigger words, the inputs and outputs, the rules that must

never be broken, and whether scripts are needed. Do not ask about the folder
name or the layout; decide those with the rules above.

### Output

Create the folder with `SKILL.md` and any `references/`, `scripts/, evals/` or `assets/`
files. Then report:

- the path of the new skill folder;
- the `name` and the full `description`;
- the test result (what the agent did with and without the skill);
- anything left open, such as a rule the user must confirm.

Keep the report short. Do not paste the whole `SKILL.md` into chat.
