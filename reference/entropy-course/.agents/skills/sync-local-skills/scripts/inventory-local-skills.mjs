#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";

const options = parseArgs(process.argv.slice(2));

if (options.help) {
  printUsage();
  process.exit(0);
}

const repoRoot = path.resolve(options.repo ?? process.cwd());
const skillsDir = path.join(repoRoot, ".agents", "skills");
const agentsPath = path.resolve(options.agents ?? path.join(repoRoot, "AGENTS.md"));

const skills = readLocalSkills(skillsDir);

if (options.checkAgents) {
  checkAgentsCoverage(agentsPath, skills);
} else if (options.json) {
  process.stdout.write(`${JSON.stringify({ repoRoot, skills }, null, 2)}\n`);
} else {
  printMarkdownInventory(repoRoot, skills);
}

function parseArgs(args) {
  const parsed = {
    agents: undefined,
    checkAgents: false,
    help: false,
    json: false,
    repo: undefined,
  };

  for (let index = 0; index < args.length; index += 1) {
    const arg = args[index];

    if (arg === "--agents") {
      parsed.agents = readOptionValue(args, index, arg);
      index += 1;
    } else if (arg === "--check-agents") {
      parsed.checkAgents = true;
    } else if (arg === "--help" || arg === "-h") {
      parsed.help = true;
    } else if (arg === "--json") {
      parsed.json = true;
    } else if (arg === "--markdown") {
      parsed.json = false;
    } else if (arg === "--repo") {
      parsed.repo = readOptionValue(args, index, arg);
      index += 1;
    } else {
      throw new Error(`Unknown argument: ${arg}`);
    }
  }

  return parsed;
}

function readOptionValue(args, index, arg) {
  const value = args[index + 1];

  if (!value || value.startsWith("--")) {
    throw new Error(`${arg} requires a value`);
  }

  return value;
}

function readLocalSkills(directory) {
  if (!fs.existsSync(directory)) {
    throw new Error(`Skills directory not found: ${directory}`);
  }

  return fs
    .readdirSync(directory, { withFileTypes: true })
    .filter((entry) => entry.isDirectory())
    .map((entry) => {
      const skillPath = path.join(directory, entry.name, "SKILL.md");

      if (!fs.existsSync(skillPath)) {
        return null;
      }

      const skillText = fs.readFileSync(skillPath, "utf8");
      const frontmatter = parseFrontmatter(skillText);

      return {
        description: frontmatter.description ?? "",
        directory: entry.name,
        name: frontmatter.name ?? entry.name,
        skillPath,
        token: `\${${entry.name}}`,
      };
    })
    .filter(Boolean)
    .sort((left, right) => left.directory.localeCompare(right.directory));
}

function parseFrontmatter(text) {
  const lines = text.split(/\r?\n/);

  if (lines[0] !== "---") {
    return {};
  }

  const endIndex = lines.findIndex((line, index) => index > 0 && line === "---");

  if (endIndex === -1) {
    return {};
  }

  const result = {};
  const frontmatterLines = lines.slice(1, endIndex);

  for (let index = 0; index < frontmatterLines.length; index += 1) {
    const line = frontmatterLines[index];
    const match = line.match(/^([A-Za-z0-9_-]+):\s*(.*)$/);

    if (!match) {
      continue;
    }

    const [, key, rawValue] = match;

    if (key !== "name" && key !== "description") {
      continue;
    }

    const valueLines = [rawValue];

    while (index + 1 < frontmatterLines.length && /^\s+\S/.test(frontmatterLines[index + 1])) {
      index += 1;
      valueLines.push(frontmatterLines[index].trim());
    }

    result[key] = normalizeYamlScalar(valueLines.join(" "));
  }

  return result;
}

function normalizeYamlScalar(value) {
  const trimmed = value.trim();
  const unquoted =
    (trimmed.startsWith('"') && trimmed.endsWith('"')) ||
    (trimmed.startsWith("'") && trimmed.endsWith("'"))
      ? trimmed.slice(1, -1)
      : trimmed;

  return toAscii(unquoted).replace(/\\"/g, '"').replace(/''/g, "'").replace(/\s+/g, " ").trim();
}

function toAscii(value) {
  return value
    .replace(/[“”]/g, '"')
    .replace(/[‘’]/g, "'")
    .replace(/[–—]/g, "-")
    .replace(/\u00a0/g, " ");
}

function printMarkdownInventory(root, skillList) {
  process.stdout.write(
    `Found ${skillList.length} local skills under ${path.join(root, ".agents", "skills")}.\n\n`,
  );
  process.stdout.write("| Token | Directory | Description |\n");
  process.stdout.write("| --- | --- | --- |\n");

  for (const skill of skillList) {
    process.stdout.write(
      `| \`${skill.token}\` | \`${skill.directory}\` | ${escapeMarkdownTable(skill.description || "No description")} |\n`,
    );
  }
}

function escapeMarkdownTable(value) {
  return value.replace(/\|/g, "\\|");
}

function checkAgentsCoverage(filePath, skillList) {
  if (!fs.existsSync(filePath)) {
    throw new Error(`AGENTS.md not found: ${filePath}`);
  }

  const agentsText = fs.readFileSync(filePath, "utf8");
  const missing = skillList.filter((skill) => !agentsText.includes(skill.token));

  if (missing.length > 0) {
    for (const skill of missing) {
      process.stdout.write(`MISSING ${skill.token}\n`);
    }

    process.exitCode = 1;
    return;
  }

  process.stdout.write(`All ${skillList.length} local skill tokens are listed in ${filePath}\n`);
}

function printUsage() {
  process.stdout.write(
    `Usage: node inventory-local-skills.mjs [--repo <path>] [--markdown|--json] [--check-agents] [--agents <path>]\n`,
  );
}
