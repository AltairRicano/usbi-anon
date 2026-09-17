# Multi-etapa (M3.2/M3.8): el compilador de Go y el código fuente nunca
# viajan a producción, solo los dos binarios estáticos que produce el build.
FROM golang:1.22-bookworm AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /out/usbi-anon-backend .
RUN CGO_ENABLED=0 go build -trimpath -o /out/usbictl ./cmd/usbictl

# distroless: sin shell, sin gestor de paquetes, superficie mínima — los
# binarios de arriba son estáticos (CGO_ENABLED=0) y no necesitan nada más.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/usbi-anon-backend /usbi-anon-backend
COPY --from=build /out/usbictl /usbictl

EXPOSE 8088
ENTRYPOINT ["/usbi-anon-backend"]
