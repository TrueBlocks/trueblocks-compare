FROM golang:1.22-alpine as builder
WORKDIR /src
COPY . .
RUN go build -o compare .

FROM --platform=linux/x86_64 trueblocks/core:v3.0.0-release
COPY --from=builder /src .
CMD ./compare --datadir /host /host/addresses.txt

# docker run --rm --platform linux/amd64 --network host --mount type=bind,src=./host/,dst=/host --mount type=bind,src=./tbconfig/trueBlocks.toml,dst=/root/.local/share/trueblocks/trueBlocks.toml,ro --mount type=bind,src=./unchained,dst=/unchained compare
