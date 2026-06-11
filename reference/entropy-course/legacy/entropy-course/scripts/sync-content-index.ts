#!/usr/bin/env bun
/**
 * sync-content-index.ts
 * ---------------------------------------------------------
 * Reads content/manifest.json and upserts track / stage /
 * lesson rows in Postgres via Drizzle. Short-circuits on
 * unchanged track content_hash so re-runs are near-instant.
 *
 * Optionally prunes rows for tracks/stages/lessons that have
 * been removed from the manifest (pass --prune).
 *
 * Run with:
 *   bun run scripts/sync-content-index.ts
 *   bun run scripts/sync-content-index.ts --prune
 */

import { promises as fs } from "node:fs";
import path from "node:path";
import { eq, inArray, notInArray, and } from "drizzle-orm";
import { db } from "@entropy-course/db";
import { track, stage, lesson } from "@entropy-course/db/schema/content";

// ---------- paths ----------
const ROOT = path.resolve(import.meta.dir, "..");
const MANIFEST_PATH = path.join(ROOT, "content/manifest.json");

// ---------- manifest types (mirror build-manifest.ts) ----------
type ManifestLesson = {
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
type ManifestStage = {
  id: string;
  slug: string;
  title: string;
  order_index: number;
  lessons: ManifestLesson[];
};
type ManifestTrack = {
  id: string;
  slug: string;
  title: string;
  summary: string | null;
  order_index: number;
  version: string;
  content_hash: string;
  status: "draft" | "review" | "published";
  stages: ManifestStage[];
};
type Manifest = { generated_at: string; tracks: ManifestTrack[] };

// ---------- helpers ----------
async function readManifest(): Promise<Manifest> {
  return JSON.parse(await fs.readFile(MANIFEST_PATH, "utf8"));
}

async function syncTrack(t: ManifestTrack): Promise<"unchanged" | "synced"> {
  // Short-circuit: if the track-level content_hash matches what's already in
  // the DB, nothing inside the track has changed since the last sync.
  const existing = await db
    .select({ id: track.id, contentHash: track.contentHash })
    .from(track)
    .where(eq(track.id, t.id))
    .limit(1);

  if (existing[0]?.contentHash === t.content_hash) {
    return "unchanged";
  }

  const publishedAt = t.status === "published" ? new Date() : null;

  await db
    .insert(track)
    .values({
      id: t.id,
      slug: t.slug,
      title: t.title,
      summary: t.summary,
      orderIndex: t.order_index,
      version: t.version,
      contentHash: t.content_hash,
      publishedAt,
    })
    .onConflictDoUpdate({
      target: track.id,
      set: {
        slug: t.slug,
        title: t.title,
        summary: t.summary,
        orderIndex: t.order_index,
        version: t.version,
        contentHash: t.content_hash,
        publishedAt,
        updatedAt: new Date(),
      },
    });

  for (const s of t.stages) {
    await db
      .insert(stage)
      .values({
        id: s.id,
        trackId: t.id,
        slug: s.slug,
        title: s.title,
        orderIndex: s.order_index,
      })
      .onConflictDoUpdate({
        target: stage.id,
        set: {
          trackId: t.id,
          slug: s.slug,
          title: s.title,
          orderIndex: s.order_index,
        },
      });

    for (const l of s.lessons) {
      await db
        .insert(lesson)
        .values({
          id: l.id,
          stageId: s.id,
          slug: l.slug,
          title: l.title,
          orderIndex: l.order_index,
          estimatedMinutes: l.estimated_minutes,
          contentHash: l.content_hash,
          videoProvider: l.video_provider,
          videoAssetId: l.video_asset_id,
          objectives: l.objectives,
        })
        .onConflictDoUpdate({
          target: lesson.id,
          set: {
            stageId: s.id,
            slug: l.slug,
            title: l.title,
            orderIndex: l.order_index,
            estimatedMinutes: l.estimated_minutes,
            contentHash: l.content_hash,
            videoProvider: l.video_provider,
            videoAssetId: l.video_asset_id,
            objectives: l.objectives,
          },
        });
    }
  }

  return "synced";
}

/** Delete rows that no longer appear in the manifest. */
async function prune(manifest: Manifest) {
  const trackIds = manifest.tracks.map((t) => t.id);
  const stageIds = manifest.tracks.flatMap((t) => t.stages.map((s) => s.id));
  const lessonIds = manifest.tracks.flatMap((t) =>
    t.stages.flatMap((s) => s.lessons.map((l) => l.id)),
  );

  // Delete lessons first (FK), then stages, then tracks.
  const deletedLessons = await db
    .delete(lesson)
    .where(lessonIds.length > 0 ? notInArray(lesson.id, lessonIds) : undefined)
    .returning({ id: lesson.id });
  const deletedStages = await db
    .delete(stage)
    .where(stageIds.length > 0 ? notInArray(stage.id, stageIds) : undefined)
    .returning({ id: stage.id });
  const deletedTracks = await db
    .delete(track)
    .where(trackIds.length > 0 ? notInArray(track.id, trackIds) : undefined)
    .returning({ id: track.id });

  if (deletedLessons.length || deletedStages.length || deletedTracks.length) {
    console.log(
      `  pruned: ${deletedTracks.length} track(s), ${deletedStages.length} stage(s), ${deletedLessons.length} lesson(s)`,
    );
  }
}

// ---------- main ----------
async function main() {
  const shouldPrune = process.argv.includes("--prune");
  const manifest = await readManifest();

  let synced = 0;
  let unchanged = 0;
  for (const t of manifest.tracks) {
    const result = await syncTrack(t);
    if (result === "synced") {
      synced++;
      console.log(`✓ ${t.slug} synced (${t.stages.length} stages)`);
    } else {
      unchanged++;
      console.log(`= ${t.slug} unchanged`);
    }
  }

  if (shouldPrune) await prune(manifest);

  console.log(
    `done — ${synced} synced, ${unchanged} unchanged, ${manifest.tracks.length} total`,
  );
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
