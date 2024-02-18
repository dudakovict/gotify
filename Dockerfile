FROM golang:1.21-alpine3.18 as builder
ENV CGO_ENABLED 0
ARG BUILD_REF

WORKDIR /app

COPY . .

RUN go build -ldflags "-X main.build=${BUILD_REF}" -o gotify ./cmd

FROM alpine:3.18
ARG BUILD_DATE
ARG BUILD_REF

WORKDIR /app

COPY --from=builder /app/gotify .
COPY app.env .
COPY start.sh .
COPY wait-for.sh .
COPY platform/migrations ./platform/migrations

RUN chmod +x ./start.sh
RUN chmod +x ./wait-for.sh

EXPOSE 3000

CMD [ "/app/gotify" ]
ENTRYPOINT [ "/app/start.sh" ]

LABEL org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.title="gotify-api" \
      org.opencontainers.image.authors="Timon Dudaković <dudakovict@gmail.com>" \
      org.opencontainers.image.source="https://github.com/dudakovict/gotify" \
      org.opencontainers.image.revision="${BUILD_REF}" \
      org.opencontainers.image.vendor="dudakovict"