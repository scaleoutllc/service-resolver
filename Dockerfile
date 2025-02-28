# This base image is preferred over 1.24 because it has no known CVEs
FROM golang:1.23.6-alpine3.21 AS build
RUN apk --no-cache add make
WORKDIR /app
COPY . .
RUN make

# No base image necessary, no base vulnerabilities to sweat
FROM scratch
COPY --from=build /app/dist/service-resolver /service-resolver
ENTRYPOINT ["/service-resolver"]