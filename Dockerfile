FROM golang:1.27-trixie AS build

# Runtime dependencies
RUN apt-get update \
    && apt-get install -y --no-install-recommends arp-scan libcap2-bin \
    && mkdir -p /out/lib \
    && ldd $(which arp-scan) \
       | awk '{for(i=1;i<=NF;i++) if($i=="=>") print $(i+1)}' \
       | xargs -I '{}' cp -v '{}' /out/lib \
    && setcap cap_net_raw+p $(which arp-scan)

# Cache Go dependencies
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

# Build the application
COPY . .
RUN go build -o /app/arpmon .

FROM gcr.io/distroless/base-debian13

COPY --from=build /out/lib /usr/lib
COPY --from=build /usr/sbin/arp-scan /usr/bin/arp-scan
COPY --from=build /app/arpmon /app/arpmon

WORKDIR /app
USER nonroot

CMD ["./arpmon"]

