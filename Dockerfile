FROM golang:alpine as build

RUN apk --no-cache add ca-certificates

WORKDIR /build
COPY . .
RUN go mod download
ENV CGO_ENABLED=0
RUN cd /build/ && go build -a -o /build/cfddns

ARG version
RUN echo "$version" > /etc/program-version
RUN chattr +i /etc/program-version

# Create the output container from the built image.
FROM scratch
COPY --from=build /build/cfddns /cfddns
COPY --from=build /etc/ssl/certs/ /etc/ssl/certs
COPY --from=build /etc/program-version /etc/program-version

ENTRYPOINT [ "/cfddns" ]