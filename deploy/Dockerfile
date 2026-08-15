# Gerege SSO's own image.
#
# It carries two binaries: this product's server — the platform plus the four
# SSO modules — and the platform's own migration runner, built from the exact
# core version go.mod pins. The migrations themselves come from the same place,
# out of the module cache, so the schema an image applies is the schema the code
# in it was written against. Taking them from anywhere else is how a deployment
# ends up one release ahead of its own database.

FROM golang:1.26-alpine AS build

RUN apk add --no-cache git ca-certificates

WORKDIR /src

# Dependencies first, so a change to our own code does not re-download the core.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# What this build tells an app store it is.
#
# Derived from the core version go.mod pins rather than passed in, because it
# is not a choice: PlatformVersion is what every manifest's
# `"platform": ">=1.6.0"` is checked against, and the honest answer is the
# version of the platform actually compiled into this image. Left to a build
# argument it defaults to 0.0.0 the first time a deploy forgets to pass one,
# and the instance refuses every manifest it ships with — it cannot read its
# own catalogue, so it will not start.
RUN core=$(go list -m -f '{{.Version}}' github.com/gerege-systems/open-gerege-nexus/backend) \
    && version=${core#v} \
    && echo "building as platform ${version}" \
    && CGO_ENABLED=0 GOOS=linux go build -trimpath \
    -ldflags="-w -s -X github.com/gerege-systems/open-gerege-nexus/backend/internal/platform.PlatformVersion=${version}" \
    -o /bin/sso .

# The core's migration runner, from the version this product depends on. It is
# a package in that module, so `go build` resolves it like any other import.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-w -s" \
    -o /bin/migrate github.com/gerege-systems/open-gerege-nexus/backend/cmd/migrate

# And the migration files, which are data in the module rather than code. The
# path carries the version, so this reads it back rather than guessing.
RUN core=$(go list -m -f '{{.Dir}}' github.com/gerege-systems/open-gerege-nexus/backend) \
    && mkdir -p /out/db && cp -r "$core/db/migrations" /out/db/migrations \
    && chmod -R u+w /out/db

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata wget \
    && adduser -D -u 10001 sso

WORKDIR /app

COPY --from=build /bin/sso     /app/sso
COPY --from=build /bin/migrate /app/migrate
COPY --from=build /out/db      /app/db
# And this product's own migrations, which are not the core's. A table only the
# four SSO modules have any use for does not belong in a migration every
# deployment of the platform runs, so this history is this repository's, applied
# with MIGRATIONS_DIR and a MIGRATIONS_TABLE of its own — the pair the core's
# cmd/migrate grew for exactly this.
COPY db/migrations /app/db/sso
# This product's catalogue: the four SSO modules and the three platform apps they
# sit beside.
COPY catalog /app/catalog

USER sso
EXPOSE 8080

CMD ["/app/sso"]
