# to test build:
# docker build -t tmp -f eval/go/Dockerfile .
FROM gcr.io/cloudeval-255302/eval-multi-base as build
WORKDIR /src/eval
COPY eval/ /src/eval
COPY defs/ /src/defs

RUN go build -o codeeval .

FROM gcr.io/cloudeval-255302/eval-multi-base

COPY --from=build /src/eval/codeeval /
CMD ["/codeeval"]
