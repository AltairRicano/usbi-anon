# Multi-etapa (M3.2/M3.8): compila dist/ con Node en una etapa y copia solo
# el resultado a una imagen nginx — al servidor no llega ni el código fuente
# ni node_modules, y el pico de memoria de `vite build` (~600-700 MB, con
# Phaser) ocurre en la máquina que construye, nunca en el contenedor final
# ni en el VPS de 1 GB de RAM de referencia.
FROM node:20-bookworm AS build
WORKDIR /src

COPY package*.json ./
COPY packages/engine/package*.json packages/engine/
COPY packages/schema/package*.json packages/schema/
RUN npm ci

COPY . .
RUN npm run build:engine
RUN npm run build

FROM nginx:1.27-alpine
COPY --from=build /src/dist /usr/share/nginx/html
COPY deploy/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
