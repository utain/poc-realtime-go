FROM --platform=amd64 golang:1.18 as builder
RUN apt install git gcc make
WORKDIR /realtime
ENV CGO_ENABLED=0
ADD ./go.mod .
ADD ./go.sum .
RUN go mod download

ADD . .

RUN make

FROM --platform=amd64 alpine as realtime

# RUN apk --no-cache add ca-certificates bash

COPY --from=builder /realtime/dist/realtime /bin/realtime

EXPOSE 6000
CMD /bin/realtime -p 6000 -t 0
