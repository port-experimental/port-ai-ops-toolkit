FROM golang:1.22-alpine AS build
WORKDIR /app
COPY . ./
RUN CGO_ENABLED=0 GO111MODULE=on go build -trimpath -ldflags="-s -w" -o /out/copilot-worker ./main.go

FROM gcr.io/distroless/static:nonroot
COPY --from=build /out/copilot-worker /copilot-worker
USER nonroot:nonroot
ENTRYPOINT ["/copilot-worker"]
