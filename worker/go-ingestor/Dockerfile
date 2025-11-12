FROM golang:1.22-alpine AS build
WORKDIR /app
COPY . ./
RUN CGO_ENABLED=0 GO111MODULE=on go build -trimpath -ldflags="-s -w" -o /out/ingest ./main.go

FROM gcr.io/distroless/static:nonroot
COPY --from=build /out/ingest /ingest
USER nonroot:nonroot
ENTRYPOINT ["/ingest"]
