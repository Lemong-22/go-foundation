# syntax=docker/dockerfile:1
FROM oven/bun:1.3.1-alpine AS build
WORKDIR /app

COPY package.json bun.lock turbo.json tsconfig.json ./
COPY apps ./apps
COPY packages ./packages

RUN bun install --frozen-lockfile
RUN bun run --cwd apps/server build

FROM oven/bun:1.3.1-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production

COPY --from=build /app/package.json /app/bun.lock ./
COPY --from=build /app/node_modules ./node_modules
COPY --from=build /app/apps/server ./apps/server
COPY --from=build /app/packages ./packages

EXPOSE 3000
CMD ["bun", "run", "--cwd", "apps/server", "start"]
