# iron/go is the alpine image with only ca-certificates added
FROM iron/go

# Now just add the binary
ADD bin/ingress-validator /

ENTRYPOINT ["./ingress-validator"]