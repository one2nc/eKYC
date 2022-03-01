##
##Build
##
FROM golang:1.17.6-alpine AS builder

WORKDIR /eKYC
COPY go.mod ./
COPY go.sum ./

COPY . ./

RUN go build -o face_match_worker.o ./cmd/face_match_worker

##
##Deploy
##
FROM alpine
COPY --from=builder /eKYC/face_match_worker.o .
COPY --from=builder /eKYC/.env .
EXPOSE 8080

ENTRYPOINT ["./face_match_worker.o" ]