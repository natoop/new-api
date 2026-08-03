FROM golang:1.26.1-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
# relaykit is referenced by the root module's replace directive.
COPY relaykit/go.mod ./relaykit/go.mod
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/affinity-admin ./extensions/channel-affinity-admin/cmd/affinity-admin

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/affinity-admin /affinity-admin
USER nonroot:nonroot
EXPOSE 8089
ENTRYPOINT ["/affinity-admin"]
