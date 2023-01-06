## Build
FROM golang:1.19.3-bullseye AS build

WORKDIR /app

ADD . .

RUN make all

## Deploy
FROM gcr.io/distroless/base-debian11
WORKDIR /

COPY --from=build /app/bin/admission-controller /admission-controller

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/admission-controller"]
