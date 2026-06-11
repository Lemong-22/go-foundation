#!/usr/bin/env bun
/**
 * build-manifest.ts
 * ---------------------------------------------------------
 * Walks /content/tracks, parses track.yaml / stage.yaml and
 * lesson MDX frontmatter, hashes each lesson's sources, and
 * emits /content/manifest.json.
 *
 * Run with:
 *   bun run scripts/build-manifest.ts
 *
 * Required deps (add to root package.json):
 *   yaml         — YAML parser
 *   gray-matter  — MDX frontmatter parser
 */

import { createHash } from "node:crypto";
import { promises as fs } from "node:fs";
import path from "node:path";
import { parse as parseYaml } from "yaml";
import matter from "gray-matter";

// ---------- paths ----------
const ROOT = path.resolve(import.meta.dir, "..");
const CONTENT_DIR = path.join(ROOT, "content");
const TRACKS_DIR = path.join(CONTENT_DIR, "tracks");
const MANIFEST_PATH = path.join(CONTENT_DIR, "manifest.json");

// ---------- types ----------
type LessonManifest = {
  id: string;
  slug: string;
  title: string;
  order_index: number;
  estimated_minutes: number | null;
  content_hash: string;
  video_provider: string | null;
  video_asset_id: string | null;
  objectives: string[];
};

type StageManifest = {
  id: string;
  slug: string;
  title: string;
  order_index: number;
  lessons: LessonManifest[];
};

type TrackManifest = {
  id: string;
  slug: string;
  title: string;
  summary: string | null;
  order_index: number;
  version: string;
  content_hash: string;
  status: "draft" | "review" | "published";
  stages: StageManifest[];
};

type Manifest = {
  generated_at: string;
  tracks: TrackManifest[];
};

// ---------- hashing ----------
async function hashFile(filePath: string): Promise<string> {
  const buf = await fs.readFile(filePath);
  return createHash("sha256").update(buf).digest("hex");
}

/** Stable hash of a directory: sorted relative paths + file bytes. */
async function hashDirectory(dir: string): Promise<string> {
  const hash = createHash("sha256");
  const files: string[] = [];

  async function walk(current: string) {
    const items = await fs.readdir(current, { withFileTypes: true });
    items.sort((a, b) => a.name.localeCompare(b.name));
    for (const item of items) {
      const full = path.join(current, item.name);
      if (item.isDirectory()) await walk(full);
      else if (item.isFile()) files.push(full);
    }
  }

  try {
    await walk(dir);
  } catch {
    // missing directory → empty hash contribution
  }

  for (const file of files) {
    hash.update(path.relative(dir, file));
    hash.update("\0");
    hash.update(await fs.readFile(file));
    hash.update("\0");
  }
  return hash.digest("hex");
}

// ---------- builders ----------
async function buildLesson(
  stageDir: string,
  lessonSlug: string,
  fallbackOrder: number,
  trackSlug: string,
  stageSlug: string,
): Promise<LessonManifest> {
  const mdxPath = path.join(stageDir, `${lessonSlug}.mdx`);
  const assetsDir = path.join(stageDir, "assets", lessonSlug);

  const raw = await fs.readFile(mdxPath, "utf8");
  const { data: fm } = matter(raw);

  // hash = MDX bytes + per-lesson asset folder bytes
  const h = createHash("sha256");
  h.update(await hashFile(mdxPath));
  try {
    await fs.access(assetsDir);
    h.update(await hashDirectory(assetsDir));
  } catch {
    // no per-lesson assets — fine
  }

  return {
    id: `${trackSlug}:${stageSlug}:${lessonSlug}`,
    slug: lessonSlug,
    title: String(fm.title ?? lessonSlug),
    order_index: Number(fm.order_index ?? fallbackOrder),
    estimated_minutes:
      typeof fm.estimated_minutes === "number" ? fm.estimated_minutes : null,
    content_hash: h.digest("hex"),
    video_provider: fm.video?.provider ?? null,
    video_asset_id: fm.video?.asset_id ?? null,
    objectives: Array.isArray(fm.objectives) ? fm.objectives.map(String) : [],
  };
}

async function buildStage(
  trackDir: string,
  stageSlug: string,
  fallbackOrder: number,
  trackSlug: string,
): Promise<StageManifest> {
  const stageDir = path.join(trackDir, stageSlug);
  const stageYaml = parseYaml(
    await fs.readFile(path.join(stageDir, "stage.yaml"), "utf8"),
  ) ?? {};

  const lessonSlugs: string[] = Array.isArray(stageYaml.lessons)
    ? stageYaml.lessons
    : [];

  const lessons: LessonManifest[] = [];
  for (let i = 0; i < lessonSlugs.length; i++) {
    lessons.push(
      await buildLesson(stageDir, lessonSlugs[i], i + 1, trackSlug, stageSlug),
    );
  }

  return {
    id: `${trackSlug}:${stageSlug}`,
    slug: stageSlug,
    title: String(stageYaml.title ?? stageSlug),
    order_index: Number(stageYaml.order_index ?? fallbackOrder),
    lessons,
  };
}

async function buildTrack(trackSlug: string): Promise<TrackManifest> {
  const trackDir = path.join(TRACKS_DIR, trackSlug);
  const trackYaml = parseYaml(
    await fs.readFile(path.join(trackDir, "track.yaml"), "utf8"),
  ) ?? {};

  const stageSlugs: string[] = Array.isArray(trackYaml.stages)
    ? trackYaml.stages
    : [];

  const stages: StageManifest[] = [];
  for (let i = 0; i < stageSlugs.length; i++) {
    stages.push(await buildStage(trackDir, stageSlugs[i], i + 1, trackSlug));
  }

  return {
    id: trackSlug,
    slug: trackSlug,
    title: String(trackYaml.title ?? trackSlug),
    summary: trackYaml.summary ?? null,
    order_index: Number(trackYaml.order_index ?? 0),
    version: String(trackYaml.version ?? "0.0.0"),
    content_hash: await hashDirectory(trackDir),
    status: (trackYaml.status ?? "draft") as TrackManifest["status"],
    stages,
  };
}

// ---------- main ----------
async function main() {
  const entries = await fs.readdir(TRACKS_DIR, { withFileTypes: true });
  const trackSlugs = entries
    .filter((e) => e.isDirectory())
    .map((e) => e.name)
    .sort();

  const tracks: TrackManifest[] = [];
  for (const slug of trackSlugs) {
    try {
      tracks.push(await buildTrack(slug));
    } catch (err) {
      console.error(`✗ track "${slug}" failed:`, (err as Error).message);
      process.exitCode = 1;
    }
  }

  tracks.sort((a, b) => a.order_index - b.order_index);

  const manifest: Manifest = {
    generated_at: new Date().toISOString(),
    tracks,
  };

  await fs.writeFile(MANIFEST_PATH, JSON.stringify(manifest, null, 2) + "\n");

  const lessonCount = tracks.reduce(
    (n, t) => n + t.stages.reduce((m, s) => m + s.lessons.length, 0),
    0,
  );
  const stageCount = tracks.reduce((n, t) => n + t.stages.length, 0);
  console.log(`✓ wrote ${path.relative(ROOT, MANIFEST_PATH)}`);
  console.log(
    `  ${tracks.length} track(s) · ${stageCount} stage(s) · ${lessonCount} lesson(s)`,
  );
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
