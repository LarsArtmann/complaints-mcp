FROM gcr.io/distroless/static-debian13:nonroot

COPY complaints-mcp /complaints-mcp

USER 65532:65532

ENTRYPOINT ["/complaints-mcp"]
