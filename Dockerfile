FROM alpine:3.11
LABEL maintainers="joshvanl"
LABEL description="cert-manager CSI Driver"

# Add util-linux to get a new version of losetup.
RUN apk add util-linux
COPY ./bin/cert-manager-csi /cert-manager-csi
ENTRYPOINT ["/cert-manager-csi"]
