FROM node:24-alpine AS build
WORKDIR /src
COPY package*.json ./
COPY apps/web/package.json apps/web/package.json
COPY apps/mobile/package.json apps/mobile/package.json
COPY packages/api-client/package.json packages/api-client/package.json
RUN npm ci
COPY apps/web apps/web
COPY packages/api-client packages/api-client
RUN npm run build --workspace web
FROM node:24-alpine
ENV NODE_ENV=production
WORKDIR /src
COPY --from=build /src ./
USER node
EXPOSE 3000
CMD ["npm","run","start","--workspace","web"]
