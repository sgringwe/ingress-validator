# iron/go is the alpine image with only ca-certificates added
FROM iron/go

WORKDIR /bin

# Now just add the binary
ADD ingress-validator ingress-validator

ENTRYPOINT ["./ingress-validator"]