import { hashPassword } from "better-auth/crypto";
import { config } from "dotenv";
import { drizzle, type NodePgDatabase } from "drizzle-orm/node-postgres";
import { Client } from "pg";

import { account, user } from "./schema/auth";

config({ path: new URL("../../../apps/server/.env", import.meta.url).pathname });
config({ path: new URL("../../../.env", import.meta.url).pathname });

interface SeedUser {
  id: string;
  name: string;
  email: string;
}

const defaultSeedPassword = "local-password-123";

const seedUsers: SeedUser[] = [
  {
    id: "seed-user-stephen",
    name: "Stephen Antoni",
    email: process.env.SEED_STEPHEN_EMAIL ?? "stephen.local@example.com",
  },
  {
    id: "seed-user-yosi",
    name: "Yosi",
    email: process.env.SEED_YOSI_EMAIL ?? "yosi.local@example.com",
  },
];

function getSeedPassword() {
  const password = process.env.SEED_PASSWORD ?? defaultSeedPassword;

  if (password.length < 8) {
    throw new Error("SEED_PASSWORD must be at least 8 characters long.");
  }

  return password;
}

function getDatabaseUrl() {
  const databaseUrl = process.env.DATABASE_URL;

  if (!databaseUrl) {
    throw new Error(
      "DATABASE_URL is missing. Copy apps/server/.env.example to apps/server/.env before running db:seed.",
    );
  }

  return databaseUrl;
}

async function upsertSeedUser(db: NodePgDatabase, seedUser: SeedUser, passwordHash: string) {
  const now = new Date();

  await db
    .insert(user)
    .values({
      id: seedUser.id,
      name: seedUser.name,
      email: seedUser.email.toLowerCase(),
      emailVerified: true,
      createdAt: now,
      updatedAt: now,
    })
    .onConflictDoUpdate({
      target: user.id,
      set: {
        name: seedUser.name,
        email: seedUser.email.toLowerCase(),
        emailVerified: true,
        updatedAt: now,
      },
    });

  await db
    .insert(account)
    .values({
      id: `${seedUser.id}-credential`,
      accountId: seedUser.id,
      providerId: "credential",
      userId: seedUser.id,
      password: passwordHash,
      createdAt: now,
      updatedAt: now,
    })
    .onConflictDoUpdate({
      target: account.id,
      set: {
        accountId: seedUser.id,
        providerId: "credential",
        userId: seedUser.id,
        password: passwordHash,
        updatedAt: now,
      },
    });
}

async function main() {
  const client = new Client({ connectionString: getDatabaseUrl() });
  await client.connect();

  try {
    const db = drizzle(client);
    const seedPassword = getSeedPassword();
    const passwordHash = await hashPassword(seedPassword);

    for (const seedUser of seedUsers) {
      await upsertSeedUser(db, seedUser, passwordHash);
      console.log(`Seeded ${seedUser.name} <${seedUser.email.toLowerCase()}>`);
    }

    console.log("\nLocal seed complete.");
    console.log(`Default password: ${seedPassword}`);
    console.log("Override with SEED_PASSWORD, SEED_STEPHEN_EMAIL, or SEED_YOSI_EMAIL when needed.");
  } finally {
    await client.end();
  }
}

main().catch((error) => {
  console.error("Failed to seed the local database:", error);
  process.exit(1);
});
