FROM golang:1.21

COPY go.* *.go /usr/src/tracklog/
COPY cmd/ /usr/src/tracklog/cmd/
COPY pkg/ /usr/src/tracklog/pkg/

ENV CGO_ENABLED 0
WORKDIR /usr/src/tracklog

RUN go build -o cmd/server/server ./cmd/server
RUN go build -o cmd/control/control ./cmd/control


FROM node:18-slim

WORKDIR /usr/src/tracklog

COPY package.json .babelrc /usr/src/tracklog/
RUN npm install

COPY css/ /usr/src/tracklog/css/
COPY js/ /usr/src/tracklog/js/
RUN npm run build


FROM scratch

COPY --from=1 /usr/src/tracklog/public /public/
COPY ./templates /templates/
COPY --from=0 /usr/src/tracklog/cmd/server/server /bin/
COPY --from=0 /usr/src/tracklog/cmd/control/control /bin/

CMD ["/bin/server"]
