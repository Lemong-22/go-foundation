# syntax=docker/dockerfile:1
FROM oven/bun:1.3.1-alpine AS build
WORKDIR /app

ARG NEXT_PUBLIC_SERVER_URL=http://localhost:3000
ENV NEXT_PUBLIC_SERVER_URL=$NEXT_PUBLIC_SERVER_URL

COPY package.json bun.lock turbo.json tsconfig.json ./
COPY apps ./apps
COPY packages ./packages

RUN bun install --frozen-lockfile
RUN bun run --cwd apps/web build

FROM oven/bun:1.3.1-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production
ENV PORT=3001
ENV HOSTNAME=0.0.0.0

COPY --from=build /app/package.json /app/bun.lock ./
COPY --from=build /app/node_modules ./node_modules
COPY --from=build /app/apps/web ./apps/web
COPY --from=build /app/packages ./packages

EXPOSE 3001
CMD ["bun", "run", "--cwd", "apps/web", "start"]
