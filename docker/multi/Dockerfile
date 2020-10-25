# to test build:
# docker build -t tmp -f eval/go/Dockerfile .
FROM gcr.io/cloudeval-255302/eval-multi-base-20_04 as build
WORKDIR /src/eval
COPY . /src/eval
COPY defs /src/eval/defs

RUN go build -o codeeval .

FROM gcr.io/cloudeval-255302/eval-multi-base-20_04

COPY --from=build /src/eval/codeeval /
CMD ["/codeeval"]
