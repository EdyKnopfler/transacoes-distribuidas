FROM golang:1.23.2

WORKDIR /go/src
ENV PATH="/go/bin:${PATH}"

CMD ["tail", "-f", "/dev/null"]