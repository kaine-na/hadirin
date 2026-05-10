# ---- Builder ----
FROM node:20-alpine AS builder

WORKDIR /app

# Install deps (lockfile-aware)
COPY package.json package-lock.json ./
RUN npm ci

# Copy source + build
COPY . .
RUN npm run build

# ---- Runtime (nginx serves static SPA) ----
FROM nginx:alpine AS runtime

# Replace default nginx config with SPA-friendly config (supports SvelteKit adapter-static fallback)
RUN rm -f /etc/nginx/conf.d/default.conf
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy static build output
COPY --from=builder /app/build /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
